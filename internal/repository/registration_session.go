// internal/repository/registration_session.go
package repository

import (
	"github.com/azmiagr/cakra-hackathon/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IRegistrationSessionRepository interface {
	GetByUserIDForUpdate(tx *gorm.DB, userID uuid.UUID) (*entity.RegistrationSession, error)
	GetByTokenHashForUpdate(tx *gorm.DB, tokenHash string) (*entity.RegistrationSession, error)
	CreateRegistrationSession(tx *gorm.DB, session *entity.RegistrationSession) error
	UpdateRegistrationSession(tx *gorm.DB, session *entity.RegistrationSession) error
}

type RegistrationSessionRepository struct {
	db *gorm.DB
}

func NewRegistrationSessionRepository(db *gorm.DB) IRegistrationSessionRepository {
	return &RegistrationSessionRepository{db: db}
}

func (r *RegistrationSessionRepository) GetByUserIDForUpdate(tx *gorm.DB, userID uuid.UUID) (*entity.RegistrationSession, error) {
	var session entity.RegistrationSession

	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", userID).
		First(&session).Error
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *RegistrationSessionRepository) GetByTokenHashForUpdate(tx *gorm.DB, tokenHash string) (*entity.RegistrationSession, error) {
	var session entity.RegistrationSession

	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("token_hash = ?", tokenHash).
		First(&session).Error
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *RegistrationSessionRepository) CreateRegistrationSession(tx *gorm.DB, session *entity.RegistrationSession) error {
	err := tx.Create(session).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *RegistrationSessionRepository) UpdateRegistrationSession(tx *gorm.DB, session *entity.RegistrationSession) error {
	err := tx.Save(session).Error
	if err != nil {
		return err
	}
	return nil
}
