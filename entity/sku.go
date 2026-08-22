package entity

import (
	"time"

	"github.com/google/uuid"
)

type SKU struct {
	SKUID     uuid.UUID `json:"sku_id" gorm:"type:varchar(36);primaryKey"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:varchar(36);not null;index;uniqueIndex:ux_user_sku_name"`
	Name      string    `json:"name" gorm:"type:varchar(255);not null;uniqueIndex:ux_user_sku_name"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	AnalysisSessions []AnalysisSession `gorm:"foreignKey:SKUID;references:SKUID;constraint:OnDelete:CASCADE"`
}
