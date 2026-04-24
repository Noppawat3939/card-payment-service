package domain

import "errors"

var (
	// common
	ErrBodyInvalid          = errors.New("invalid body request")
	ErrMissingMerchantID    = errors.New("missing merchant_id")
	ErrMissingTransactionID = errors.New("missing transaction_id")
	ErrMissingApiKey        = errors.New("missing api key")
	ErrInvalidApiKey        = errors.New("invalid api key")

	// merchants
	ErrMerchantAlreadyExists     = errors.New("merchant email already exists")
	ErrMerchantNotFound          = errors.New("merchant email not found")
	ErrMerchantStatusNotAccepted = errors.New("merchant current status not accepted")
	ErrMerchantNotActive         = errors.New("merchant not active")

	// payments
	ErrDuplicateIdempotencyKey    = errors.New("idempotency key is duplicated")
	ErrGatewayRejected            = errors.New("gateway request rejected")
	ErrInvalidGatewayRef          = errors.New("invalid gateway reference")
	ErrTokenizeCard               = errors.New("card information invalid")
	ErrTransactionNotCapturable   = errors.New("transaction status not capturable")
	ErrTransactionNotFound        = errors.New("transaction not found")
	ErrDuplicateRequest           = errors.New("duplicated request")
	ErrTransactionAlreadyVoided   = errors.New("transaction already voided")
	ErrTransactionAlreadyRefunded = errors.New("transaction already refunded")
	ErrTransactionNotRefundable   = errors.New("transaction status not refundable")

	// gateway (mock)
	ErrCardAmoutInvalid  = errors.New("amount invalid")
	ErrCardCaptureFailed = errors.New("capture failed")
	ErrCardDeclinded     = errors.New("card declined")
	ErrCardInforInvalid  = errors.New("card information invalid")
	ErrExpiredCard       = errors.New("expired_card")
	ErrInsufficientFunds = errors.New("insufficient funds")
)
