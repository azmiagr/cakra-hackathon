package entity

import (
	"time"

	"github.com/google/uuid"
)

type AnalysisUpload struct {
	AnalysisUploadID uuid.UUID `json:"analysis_upload_id" gorm:"type:varchar(36);primaryKey"`
	UserID           uuid.UUID `json:"user_id" gorm:"type:varchar(36);not null;index"`
	OriginalFilename string    `json:"original_filename" gorm:"type:varchar(255);not null"`
	StorageObjectKey string    `json:"storage_object_key" gorm:"type:varchar(512);not null;uniqueIndex"`
	FileSizeBytes    int64     `json:"file_size_bytes" gorm:"not null"`
	SHA256           string    `json:"sha256" gorm:"type:char(64);not null"`
	Status           string    `json:"status" gorm:"type:varchar(32);not null;index"`
	SKUName          string    `json:"sku_name" gorm:"type:varchar(255)"`
	ValidRowCount    int       `json:"valid_row_count" gorm:"not null;default:0"`
	ErrorRowCount    int       `json:"error_row_count" gorm:"not null;default:0"`
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	AnalysisUploadRows     []AnalysisUploadRow     `gorm:"foreignKey:AnalysisUploadID;references:AnalysisUploadID;constraint:OnDelete:CASCADE"`
	UploadValidationErrors []UploadValidationError `gorm:"foreignKey:AnalysisUploadID;references:AnalysisUploadID;constraint:OnDelete:CASCADE"`
	AnalysisSessions       []AnalysisSession       `gorm:"foreignKey:AnalysisUploadID;references:AnalysisUploadID;constraint:OnDelete:CASCADE"`
}
