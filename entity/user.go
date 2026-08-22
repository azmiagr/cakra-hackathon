package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	UserID    uuid.UUID `json:"id" gorm:"type:varchar(36);primaryKey"`
	RoleID    uuid.UUID `json:"role_id" gorm:"type:varchar(36)"`
	Email     string    `json:"email" gorm:"type:varchar(100);not null;unique"`
	Password  string    `json:"password" gorm:"type:varchar(255);not null"`
	Status    string    `json:"status" gorm:"type:enum('active','inactive');default:'inactive'"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	OtpCodes []OtpCode `gorm:"foreignKey:UserID;references:UserID;constraint:OnDelete:CASCADE"`
}
