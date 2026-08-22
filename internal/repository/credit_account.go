package repository

import (
	"errors"
	"fmt"

	"github.com/azmiagr/cakra-hackathon/entity"
	constants "github.com/azmiagr/cakra-hackathon/pkg/constant"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrInsufficientCredits         = errors.New("insufficient available credits")
	ErrInvalidCreditAmount         = errors.New("credit amount must be positive")
	ErrInsufficientReservedCredits = errors.New("insufficient reserved credits")
)

type ICreditAccountRepository interface {
	CreateCreditAccount(tx *gorm.DB, account *entity.CreditAccount) error
	CreateTransaction(tx *gorm.DB, transaction *entity.CreditTransaction) error
	GetByUserID(tx *gorm.DB, userID uuid.UUID) (*entity.CreditAccount, error)
	GetByUserIDForUpdate(tx *gorm.DB, userID uuid.UUID) (*entity.CreditAccount, error)
	ReserveAnalysisCredit(tx *gorm.DB, userID uuid.UUID, cost int) error
	CompleteAnalysisDebit(tx *gorm.DB, userID, analysisSessionID uuid.UUID, cost int) error
	ReleaseAnalysisCredit(tx *gorm.DB, userID uuid.UUID, cost int) error
}

type CreditAccountRepository struct {
	db *gorm.DB
}

func NewCreditAccountRepository(db *gorm.DB) ICreditAccountRepository {
	return &CreditAccountRepository{db: db}
}

func (r *CreditAccountRepository) CreateCreditAccount(tx *gorm.DB, account *entity.CreditAccount) error {
	err := tx.Create(account).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *CreditAccountRepository) CreateTransaction(tx *gorm.DB, transaction *entity.CreditTransaction) error {
	err := tx.Create(transaction).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *CreditAccountRepository) GetByUserIDForUpdate(tx *gorm.DB, userID uuid.UUID) (*entity.CreditAccount, error) {
	var account entity.CreditAccount
	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", userID).
		First(&account).Error
	if err != nil {
		return nil, err
	}

	return &account, nil
}

func (r *CreditAccountRepository) GetByUserID(tx *gorm.DB, userID uuid.UUID) (*entity.CreditAccount, error) {
	var account entity.CreditAccount
	err := tx.Where("user_id = ?", userID).First(&account).Error
	if err != nil {
		return nil, err
	}

	return &account, nil
}

func (r *CreditAccountRepository) ReserveAnalysisCredit(tx *gorm.DB, userID uuid.UUID, cost int) error {
	if cost <= 0 {
		return ErrInvalidCreditAmount
	}

	account, err := r.GetByUserIDForUpdate(tx, userID)
	if err != nil {
		return err
	}
	if account.Balance-account.ReservedCredits < cost {
		return ErrInsufficientCredits
	}

	err = tx.Model(&entity.CreditAccount{}).
		Where("credit_account_id = ?", account.CreditAccountID).
		Update("reserved_credits", account.ReservedCredits+cost).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *CreditAccountRepository) CompleteAnalysisDebit(tx *gorm.DB, userID, analysisSessionID uuid.UUID, cost int) error {
	if cost <= 0 {
		return ErrInvalidCreditAmount
	}

	account, err := r.GetByUserIDForUpdate(tx, userID)
	if err != nil {
		return err
	}

	idempotencyKey := analysisDebitIdempotencyKey(analysisSessionID)
	var existing entity.CreditTransaction
	err = tx.Where("idempotency_key = ?", idempotencyKey).First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if account.ReservedCredits < cost {
		return ErrInsufficientReservedCredits
	}
	if account.Balance < cost {
		return ErrInsufficientCredits
	}

	newBalance := account.Balance - cost
	err = tx.Model(&entity.CreditAccount{}).
		Where("credit_account_id = ?", account.CreditAccountID).
		Updates(map[string]any{
			"balance":          newBalance,
			"reserved_credits": account.ReservedCredits - cost,
		}).Error
	if err != nil {
		return err
	}

	return r.CreateTransaction(tx, &entity.CreditTransaction{
		CreditTransactionID: uuid.New(),
		CreditAccountID:     account.CreditAccountID,
		AnalysisSessionID:   &analysisSessionID,
		Type:                constants.CreditTransactionAnalysisDebit,
		Amount:              -cost,
		BalanceAfter:        newBalance,
		IdempotencyKey:      idempotencyKey,
	})
}

func (r *CreditAccountRepository) ReleaseAnalysisCredit(tx *gorm.DB, userID uuid.UUID, cost int) error {
	if cost <= 0 {
		return ErrInvalidCreditAmount
	}

	account, err := r.GetByUserIDForUpdate(tx, userID)
	if err != nil {
		return err
	}
	if account.ReservedCredits < cost {
		return ErrInsufficientReservedCredits
	}

	err = tx.Model(&entity.CreditAccount{}).
		Where("credit_account_id = ?", account.CreditAccountID).
		Update("reserved_credits", account.ReservedCredits-cost).Error
	if err != nil {
		return err
	}

	return nil
}

func analysisDebitIdempotencyKey(analysisSessionID uuid.UUID) string {
	return fmt.Sprintf("analysis:%s", analysisSessionID)
}
