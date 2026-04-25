package service

import (
	"card-payment-service/internal/domain"

	"github.com/google/uuid"
)

type ChargeInput struct {
	Amount         int64
	CardNumber     string
	Currency       string
	CVV            string
	Description    *string
	ExpiryMonth    string
	ExpiryYear     string
	IdempotencyKey string
	MerchantID     uuid.UUID
}

type CaptureInput struct {
	TransactionID uuid.UUID
	MerchantID    uuid.UUID
}

type CaptureOutput struct {
	TransactionID uuid.UUID                `json:"transaction_id"`
	Status        domain.TransactionStatus `json:"status"`
}

type ChargeOutput struct {
	TransactionID uuid.UUID                `json:"transaction_id"`
	Status        domain.TransactionStatus `json:"status"`
}

type VoidInput struct {
	TransactionID uuid.UUID
	MerchantID    uuid.UUID
}

type VoidOutput struct {
	TransactionID uuid.UUID                `json:"transaction_id"`
	Status        domain.TransactionStatus `json:"status"`
}

type RefundInput struct {
	TransactionID  uuid.UUID
	MerchantID     uuid.UUID
	IdempotencyKey uuid.UUID
}

type RefundOutput struct {
	RefundID uuid.UUID           `json:"refund_id"`
	Status   domain.RefundStatus `json:"status"`
}
