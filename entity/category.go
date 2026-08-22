package entity

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	CategoryID uuid.UUID  `json:"category_id" gorm:"type:varchar(36);primaryKey"`
	UserID     *uuid.UUID `json:"user_id,omitempty" gorm:"type:varchar(36);index;uniqueIndex:ux_user_category_name"`
	Name       string     `json:"name" gorm:"type:varchar(100);not null;uniqueIndex:ux_user_category_name"`
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}
