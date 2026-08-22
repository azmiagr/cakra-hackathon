package service

import (
	"errors"

	"github.com/azmiagr/cakra-hackathon/internal/repository"
	"github.com/azmiagr/cakra-hackathon/model"
	"github.com/azmiagr/cakra-hackathon/pkg/database/mariadb"
	appErrors "github.com/azmiagr/cakra-hackathon/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ICreditService interface {
	GetBalance(userID uuid.UUID) (*model.CreditAccountResponse, error)
}

type CreditService struct {
	db         *gorm.DB
	creditRepo repository.ICreditAccountRepository
}

func NewCreditService(creditRepo repository.ICreditAccountRepository) ICreditService {
	return &CreditService{
		db:         mariadb.Connection,
		creditRepo: creditRepo,
	}
}

func (s *CreditService) GetBalance(userID uuid.UUID) (*model.CreditAccountResponse, error) {
	account, err := s.creditRepo.GetByUserID(s.db, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.NotFound("akun kredit tidak ditemukan")
	}
	if err != nil {
		return nil, appErrors.InternalServer("gagal mengambil kredit")
	}

	return &model.CreditAccountResponse{
		Balance:          account.Balance,
		ReservedCredits:  account.ReservedCredits,
		AvailableCredits: account.Balance - account.ReservedCredits,
	}, nil
}
