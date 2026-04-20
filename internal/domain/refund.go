package domain

import (
	"time"

	"github.com/google/uuid"
)

type RefundStatus string

const (
	RefundProcessing = "processing"
	RefundCompleted  = "completed"
	RefundFailed     = "failed"
)

type Refund struct {
	ID            uuid.UUID    `json:"id"`
	TransactionID uuid.UUID    `json:"transaction_id"`
	MerchantID    uuid.UUID    `json:"merchant_id"`
	RefundRef     *string      `json:"refund_ref"`
	Amount        int64        `json:"amount"`
	Status        RefundStatus `json:"status"`
	FailedReason  *string      `json:"failed_reason"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

func (Refund) TableName() string {
	return "refunds"
}
