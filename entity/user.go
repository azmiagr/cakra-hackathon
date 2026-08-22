package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	UserID         uuid.UUID `json:"id" gorm:"type:varchar(36);primaryKey"`
	RoleID         uuid.UUID `json:"role_id" gorm:"type:varchar(36)"`
	FullName       string    `json:"full_name" gorm:"type:varchar(150);not null"`
	Email          string    `json:"email" gorm:"type:varchar(100);not null;unique"`
	Password       *string   `json:"password" gorm:"type:varchar(255)"`
	Status         string    `json:"status" gorm:"type:enum('active','inactive');default:'inactive'"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	SessionVersion int       `json:"-" gorm:"not null;default:0"`

	RegistrationSessions RegistrationSession `gorm:"foreignKey:UserID;references:UserID;constraint:OnDelete:CASCADE"`
	PasswordReset        PasswordReset       `gorm:"foreignKey:UserID;references:UserID;constraint:OnDelete:CASCADE"`
	CreditAccount        CreditAccount       `gorm:"foreignKey:UserID;references:UserID;constraint:OnDelete:CASCADE"`
	AnalysisUploads      []AnalysisUpload    `gorm:"foreignKey:UserID;references:UserID;constraint:OnDelete:CASCADE"`
	SKUs                 []SKU               `gorm:"foreignKey:UserID;references:UserID;constraint:OnDelete:CASCADE"`
	AnalysisSessions     []AnalysisSession   `gorm:"foreignKey:UserID;references:UserID;constraint:OnDelete:CASCADE"`
	OtpCodes             []OtpCode           `gorm:"foreignKey:UserID;references:UserID;constraint:OnDelete:CASCADE"`
}
