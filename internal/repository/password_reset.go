package repository

import (
	"github.com/azmiagr/cakra-hackathon/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IPasswordResetRepository interface {
	GetByUserIDForUpdate(tx *gorm.DB, userID uuid.UUID) (*entity.PasswordReset, error)
	GetByTokenHashForUpdate(tx *gorm.DB, tokenHash string) (*entity.PasswordReset, error)
	Create(tx *gorm.DB, session *entity.PasswordReset) error
	Update(tx *gorm.DB, session *entity.PasswordReset) error
}

type PasswordResetRepository struct {
	db *gorm.DB
}

func NewPasswordResetRepository(db *gorm.DB) IPasswordResetRepository {
	return &PasswordResetRepository{db: db}
}

func (r *PasswordResetRepository) GetByUserIDForUpdate(tx *gorm.DB, userID uuid.UUID) (*entity.PasswordReset, error) {
	var session entity.PasswordReset
	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", userID).
		First(&session).Error
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *PasswordResetRepository) GetByTokenHashForUpdate(tx *gorm.DB, tokenHash string) (*entity.PasswordReset, error) {
	var session entity.PasswordReset
	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("token_hash = ?", tokenHash).
		First(&session).Error
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *PasswordResetRepository) Create(tx *gorm.DB, session *entity.PasswordReset) error {
	err := tx.Create(session).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *PasswordResetRepository) Update(tx *gorm.DB, session *entity.PasswordReset) error {
	err := tx.Save(session).Error
	if err != nil {
		return err
	}

	return nil
}
