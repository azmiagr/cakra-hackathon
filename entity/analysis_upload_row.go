package entity

import (
	"time"

	"github.com/google/uuid"
)

type AnalysisUploadRow struct {
	AnalysisUploadRowID uuid.UUID `json:"analysis_upload_row_id" gorm:"type:varchar(36);primaryKey"`
	AnalysisUploadID    uuid.UUID `json:"analysis_upload_id" gorm:"type:varchar(36);not null;index;uniqueIndex:ux_upload_row_number"`
	RowNumber           int       `json:"row_number" gorm:"not null;uniqueIndex:ux_upload_row_number"`
	SaleDate            time.Time `json:"sale_date" gorm:"type:date;not null"`
	QuantitySold        int       `json:"quantity_sold" gorm:"not null"`
	SKUName             string    `json:"sku_name" gorm:"type:varchar(255);not null"`
	UnitPrice           float64   `json:"unit_price" gorm:"type:decimal(15,2);not null"`
	CreatedAt           time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt           time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
