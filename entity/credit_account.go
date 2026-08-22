package entity

import (
	"time"

	"github.com/google/uuid"
)

type CreditAccount struct {
	CreditAccountID uuid.UUID `json:"credit_account_id" gorm:"type:varchar(36);primaryKey"`
	UserID          uuid.UUID `json:"user_id" gorm:"type:varchar(36);not null;uniqueIndex"`
	Balance         int       `json:"balance" gorm:"not null;default:0"`
	ReservedCredits int       `json:"reserved_credits" gorm:"not null;default:0"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	Transactions []CreditTransaction `gorm:"foreignKey:CreditAccountID;references:CreditAccountID;constraint:OnDelete:CASCADE"`
}
