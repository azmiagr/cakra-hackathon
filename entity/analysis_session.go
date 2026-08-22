package entity

import (
	"time"

	"github.com/google/uuid"
)

type AnalysisSession struct {
	AnalysisSessionID uuid.UUID `json:"analysis_session_id" gorm:"type:varchar(36);primaryKey"`
	UserID            uuid.UUID `json:"user_id" gorm:"type:varchar(36);not null;index"`
	SKUID             uuid.UUID `json:"sku_id" gorm:"column:sku_id;type:varchar(36);not null;index"`
	AnalysisUploadID  uuid.UUID `json:"analysis_upload_id" gorm:"type:varchar(36);not null;index"`
	CurrentStock      int       `json:"current_stock" gorm:"not null"`
	LeadTimeDays      int       `json:"lead_time_days" gorm:"not null"`
	CreditCost        int       `json:"credit_cost" gorm:"not null;default:1"`
	Status            string    `json:"status" gorm:"type:varchar(32);not null;index"`
	DemandCategory    *string   `json:"demand_category" gorm:"type:varchar(32)"`
	ADIValue          *float64  `json:"adi_value"`
	CVSquaredValue    *float64  `json:"cv_squared_value"`
	FailureCode       *string   `json:"failure_code" gorm:"type:varchar(64)"`
	FailureMessage    *string   `json:"failure_message" gorm:"type:varchar(500)"`
	CreatedAt         time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	SalesHistories       []SalesHistory       `gorm:"foreignKey:AnalysisSessionID;references:AnalysisSessionID;constraint:OnDelete:CASCADE"`
	RecommendationResult RecommendationResult `gorm:"foreignKey:AnalysisSessionID;references:AnalysisSessionID;constraint:OnDelete:CASCADE"`
}
