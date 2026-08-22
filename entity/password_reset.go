package entity

import (
	"time"

	"github.com/google/uuid"
)

type PasswordReset struct {
	PasswordResetID uuid.UUID `gorm:"type:varchar(36);primaryKey"`
	UserID          uuid.UUID `gorm:"type:varchar(36);not null;uniqueIndex"`
	TokenHash       string    `gorm:"type:char(64);not null;uniqueIndex"`
	Stage           string    `gorm:"type:varchar(32);not null;index"`
	OTPCodeHash     *string   `gorm:"type:varchar(255)"`
	OTPExpiresAt    *time.Time
	AttemptCount    int        `gorm:"not null;default:0"`
	LastSentAt      *time.Time `gorm:"index"`
	ExpiresAt       time.Time  `gorm:"not null;index"`
	CreatedAt       time.Time  `gorm:"autoCreateTime"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime"`
}
