package entity

import (
	"time"

	"github.com/google/uuid"
)

type CreditTransaction struct {
	CreditTransactionID uuid.UUID  `json:"credit_transaction_id" gorm:"type:varchar(36);primaryKey"`
	CreditAccountID     uuid.UUID  `json:"credit_account_id" gorm:"type:varchar(36);not null;index"`
	AnalysisSessionID   *uuid.UUID `json:"analysis_session_id,omitempty" gorm:"type:varchar(36);index"`
	Type                string     `json:"type" gorm:"type:varchar(32);not null"`
	Amount              int        `json:"amount" gorm:"not null"`
	BalanceAfter        int        `json:"balance_after" gorm:"not null"`
	IdempotencyKey      string     `json:"idempotency_key" gorm:"type:varchar(191);not null;uniqueIndex"`
	CreatedAt           time.Time  `json:"created_at" gorm:"autoCreateTime"`
}
