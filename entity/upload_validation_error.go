package entity

import (
	"time"

	"github.com/google/uuid"
)

type UploadValidationError struct {
	UploadValidationErrorID uuid.UUID `json:"upload_validation_error_id" gorm:"type:varchar(36);primaryKey"`
	AnalysisUploadID        uuid.UUID `json:"analysis_upload_id" gorm:"type:varchar(36);not null;index"`
	RowNumber               int       `json:"row_number" gorm:"not null"`
	Code                    string    `json:"code" gorm:"type:varchar(64);not null"`
	Message                 string    `json:"message" gorm:"type:varchar(500);not null"`
	CreatedAt               time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt               time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
