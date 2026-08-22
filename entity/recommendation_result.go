package entity

import (
	"time"

	"github.com/google/uuid"
)

type RecommendationResult struct {
	RecommendationResultID uuid.UUID `json:"recommendation_result_id" gorm:"type:varchar(36);primaryKey"`
	AnalysisSessionID      uuid.UUID `json:"analysis_session_id" gorm:"type:varchar(36);not null;uniqueIndex"`
	ReorderPoint           int       `json:"reorder_point" gorm:"not null"`
	ReorderQuantity        int       `json:"reorder_quantity" gorm:"not null"`
	RiskLabel              string    `json:"risk_label" gorm:"type:varchar(32);not null"`
	RiskReason             string    `json:"risk_reason" gorm:"type:text;not null"`
	ExplanationText        string    `json:"explanation_text" gorm:"type:text;not null"`
	ForecastP50            string    `json:"forecast_p50" gorm:"type:longtext;not null"`
	ForecastP90            string    `json:"forecast_p90" gorm:"type:longtext;not null"`
	CreatedAt              time.Time `json:"created_at" gorm:"autoCreateTime"`
}
