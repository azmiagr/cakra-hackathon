package entity

import (
	"time"

	"github.com/google/uuid"
)

type SalesHistory struct {
	SalesHistoryID    uuid.UUID `json:"sales_history_id" gorm:"type:varchar(36);primaryKey"`
	AnalysisSessionID uuid.UUID `json:"analysis_session_id" gorm:"type:varchar(36);not null;index;uniqueIndex:ux_session_sale_date"`
	SaleDate          time.Time `json:"sale_date" gorm:"type:date;not null;uniqueIndex:ux_session_sale_date"`
	QuantitySold      int       `json:"quantity_sold" gorm:"not null"`
	UnitPrice         float64   `json:"unit_price" gorm:"type:decimal(15,2);not null"`
}
