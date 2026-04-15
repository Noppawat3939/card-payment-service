package domain

import (
	"time"

	"github.com/google/uuid"
)

type TransactionStatus string
type PaymentType string

const (
	DirectCharge     PaymentType = "direct_charge"
	AuthorizeCapture PaymentType = "authorize_capture"
)

const (
	TransactionStatusPending    TransactionStatus = "pending"
	TransactionStatusAuthorized TransactionStatus = "authorized"
	TransactionStatusCaptured   TransactionStatus = "captured"
	TransactionStatusFailed     TransactionStatus = "failed"
	TransactionStatusVoided     TransactionStatus = "voided"
	TransactionStatusRefunded   TransactionStatus = "refunded"
)

type Transaction struct {
	ID             uuid.UUID         `json:"id"`
	MerchantID     uuid.UUID         `json:"merchant_id"`
	GatewayRef     *string           `json:"gateway_ref"`
	PaymentType    PaymentType       `json:"payment_type"`
	CardToken      string            `json:"card_token"`
	CardLastFour   string            `json:"card_last_four"`
	CardBrand      string            `json:"card_brand"`
	Amount         int64             `json:"amount"`
	Currency       string            `json:"currency"`
	Status         TransactionStatus `json:"status"`
	Description    *string           `json:"description"`
	IdempotencyKey string            `json:"idempotency_key"`
	FailedReason   *string           `json:"failed_reason"`
	CapturedAt     *time.Time        `json:"captured_at"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

func (Transaction) TableName() string {
	return "transactions"
}
