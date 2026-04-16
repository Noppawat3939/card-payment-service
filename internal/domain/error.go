package domain

import "errors"

var (
	// common
	ErrBodyInvalid = errors.New("invalid body request")

	// merchants
	ErrMerchantAlreadyExists     = errors.New("merchant email already exists")
	ErrMerchantNotFound          = errors.New("merchant email not found")
	ErrMerchantStatusNotAccepted = errors.New("merchant current status not accepted")

	// payments
	ErrDuplicateIdempotencyKey  = errors.New("idempotency key is duplicated")
	ErrGatewayRejected          = errors.New("gateway request rejected")
	ErrTokenizeCard             = errors.New("card information invalid")
	ErrTransactionNotCapturable = errors.New("transaction status not capturable")
	ErrTransactionNotFound      = errors.New("transaction not found")
	ErrInvalidGatewayRef        = errors.New("invalid gateway reference")
)
