package dto

import (
	"card-payment-service/internal/domain"

	"github.com/google/uuid"
)

type AuthorizePaymentRequest struct {
	Amount      int64   `json:"amount" binding:"required"`
	CardNumber  string  `json:"card_number" binding:"required"`
	Currency    string  `json:"currency" binding:"required"`
	CVV         string  `json:"cvv" binding:"required"`
	Description *string `json:"description"`
	ExpiryMonth string  `json:"expiry_month" binding:"required"`
	ExpiryYear  string  `json:"expiry_year" binding:"required"`
}

type AuthorizePaymentResponse struct {
	TransactionID uuid.UUID                `json:"transaction_id"`
	Status        domain.TransactionStatus `json:"status"`
}

type CapturePaymentResponse struct {
	TransactionID uuid.UUID                `json:"transaction_id"`
	Status        domain.TransactionStatus `json:"status"`
}

type VoidPaymentResponse struct {
	TransactionID uuid.UUID                `json:"transaction_id"`
	Status        domain.TransactionStatus `json:"status"`
}

type RefundResponse struct {
	RefundID uuid.UUID           `json:"refund_id"`
	Status   domain.RefundStatus `json:"status"`
}
