package mapper

import (
	"card-payment-service/internal/handler/dto"
	"card-payment-service/internal/middleware"
	"card-payment-service/internal/service/payment"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ToChargeInput(c *gin.Context, req dto.AuthorizePaymentRequest) payment.ChargeInput {
	merchantID := c.MustGet(middleware.MerchantIDKey).(uuid.UUID)
	idemKey := c.MustGet(middleware.IdempotencyKeyContextKey).(string)

	return payment.ChargeInput{
		Amount:         req.Amount,
		CardNumber:     req.CardNumber,
		Currency:       req.Currency,
		CVV:            req.CVV,
		Description:    req.Description,
		ExpiryMonth:    req.ExpiryMonth,
		ExpiryYear:     req.ExpiryYear,
		IdempotencyKey: idemKey,
		MerchantID:     merchantID,
	}
}
