package entity

import (
	"time"

	"github.com/google/uuid"
)

type RegistrationSession struct {
	RegistrationSessionID uuid.UUID `gorm:"type:varchar(36);primaryKey"`
	UserID                uuid.UUID `gorm:"type:varchar(36);not null;uniqueIndex"`
	TokenHash             string    `gorm:"type:char(64);not null;uniqueIndex"`
	Stage                 string    `gorm:"type:varchar(32);not null;index"`
	ExpiresAt             time.Time `gorm:"not null;index"`
	CreatedAt             time.Time `gorm:"autoCreateTime"`
	UpdatedAt             time.Time `gorm:"autoUpdateTime"`
}
