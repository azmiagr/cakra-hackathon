package entity

import (
	"time"

	"github.com/google/uuid"
)

type OtpCode struct {
	OtpID        uuid.UUID  `gorm:"type:varchar(36);primaryKey"`
	UserID       uuid.UUID  `gorm:"type:varchar(36);not null;uniqueIndex"`
	CodeHash     string     `gorm:"column:code;type:varchar(255);not null"`
	ExpiresAt    time.Time  `gorm:"not null;index"`
	AttemptCount int        `gorm:"not null;default:0"`
	LastSentAt   *time.Time `gorm:"index"`
	CreatedAt    time.Time  `gorm:"autoCreateTime;not null"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime;not null"`
}
