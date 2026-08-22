// internal/repository/otp.go
package repository

import (
	"time"

	"github.com/azmiagr/cakra-hackathon/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IOtpRepository interface {
	GetOtpByUserID(tx *gorm.DB, userID uuid.UUID) (*entity.OtpCode, error)
	GetOtpByUserIDForUpdate(tx *gorm.DB, userID uuid.UUID) (*entity.OtpCode, error)
	CreateOtp(tx *gorm.DB, otp *entity.OtpCode) error
	UpdateOtp(tx *gorm.DB, otp *entity.OtpCode) error
	DeleteOtpByUserID(tx *gorm.DB, userID uuid.UUID) error
	MarkOtpSent(tx *gorm.DB, userID uuid.UUID, sentAt time.Time) error
	ClearOtpSentAt(tx *gorm.DB, userID uuid.UUID) error
}

type OtpRepository struct {
	db *gorm.DB
}

func NewOtpRepository(db *gorm.DB) IOtpRepository {
	return &OtpRepository{db: db}
}

func (r *OtpRepository) GetOtpByUserID(tx *gorm.DB, userID uuid.UUID) (*entity.OtpCode, error) {
	var otp entity.OtpCode

	err := tx.Where("user_id = ?", userID).First(&otp).Error
	if err != nil {
		return nil, err
	}

	return &otp, nil
}

func (r *OtpRepository) GetOtpByUserIDForUpdate(tx *gorm.DB, userID uuid.UUID) (*entity.OtpCode, error) {
	var otp entity.OtpCode

	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", userID).
		First(&otp).Error
	if err != nil {
		return nil, err
	}

	return &otp, nil
}

func (r *OtpRepository) CreateOtp(tx *gorm.DB, otp *entity.OtpCode) error {
	err := tx.Create(otp).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *OtpRepository) UpdateOtp(tx *gorm.DB, otp *entity.OtpCode) error {
	err := tx.Save(otp).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *OtpRepository) DeleteOtpByUserID(tx *gorm.DB, userID uuid.UUID) error {
	err := tx.Where("user_id = ?", userID).Delete(&entity.OtpCode{}).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *OtpRepository) MarkOtpSent(tx *gorm.DB, userID uuid.UUID, sentAt time.Time) error {
	err := tx.Model(&entity.OtpCode{}).
		Where("user_id = ?", userID).
		Update("last_sent_at", sentAt).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *OtpRepository) ClearOtpSentAt(tx *gorm.DB, userID uuid.UUID) error {
	err := tx.Model(&entity.OtpCode{}).
		Where("user_id = ?", userID).
		Update("last_sent_at", nil).Error
	if err != nil {
		return err
	}
	return nil
}
