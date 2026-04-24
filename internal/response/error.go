package response

import (
	"card-payment-service/internal/domain"
	"errors"
	"net/http"

	"gorm.io/gorm"
)

func mapErrStatusCode(err error) int {
	switch {
	// 404
	case errors.Is(err, domain.ErrMerchantNotFound),
		errors.Is(err, gorm.ErrRecordNotFound):
		return http.StatusNotFound
	// 406
	case errors.Is(err, domain.ErrTokenizeCard),
		errors.Is(err, domain.ErrMerchantAlreadyExists),
		errors.Is(err, domain.ErrMerchantStatusNotAccepted):
		return http.StatusNotAcceptable
	// 409
	case errors.Is(err, domain.ErrDuplicateIdempotencyKey),
		errors.Is(err, domain.ErrTransactionAlreadyVoided),
		errors.Is(err, domain.ErrDuplicateRequest):
		return http.StatusConflict

	// 422
	case errors.Is(err, domain.ErrTransactionNotCapturable),
		errors.Is(err, domain.ErrTransactionAlreadyRefunded):
		return http.StatusUnprocessableEntity
	// 402
	case errors.Is(err, domain.ErrGatewayRejected),
		errors.Is(err, domain.ErrCardInforInvalid),
		errors.Is(err, domain.ErrCardAmoutInvalid),
		errors.Is(err, domain.ErrCardCaptureFailed),
		errors.Is(err, domain.ErrCardDeclinded),
		errors.Is(err, domain.ErrExpiredCard),
		errors.Is(err, domain.ErrInsufficientFunds):
		return http.StatusPaymentRequired
	// 500 (uncontrollable)
	default:
		return http.StatusInternalServerError
	}
}
