package handler

import (
	"card-payment-service/internal/domain"
	"card-payment-service/internal/handler/dto"
	"card-payment-service/internal/middleware"
	"card-payment-service/internal/response"
	"card-payment-service/internal/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type PaymentHandler struct {
	paymentService *service.PaymentService
	log            zerolog.Logger
}

func NewPaymentHandler(paymentService *service.PaymentService, log zerolog.Logger) *PaymentHandler {
	return &PaymentHandler{paymentService, log}
}

func (h *PaymentHandler) Authorize(c *gin.Context) {
	var req dto.AuthorizePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logErr(c, err)
		response.BadRequest(c)
		return
	}

	merchantID := c.MustGet(middleware.MerchantIDKey).(uuid.UUID)
	idemKey := c.MustGet(middleware.IdempotencyKeyContextKey).(string)

	data, err := h.paymentService.Authorize(c, service.AuthorizeInput{
		Amount:         req.Amount,
		CardNumber:     req.CardNumber,
		Currency:       req.Currency,
		CVV:            req.CVV,
		Description:    req.Description,
		ExpiryMonth:    req.ExpiryMonth,
		ExpiryYear:     req.ExpiryYear,
		IdempotencyKey: idemKey,
		MerchantID:     merchantID,
	})
	if err != nil {
		statusCode := mapPaymentErrStatusCode(err)
		response.Error(c, statusCode, err.Error())
		return
	}

	response.OK(c, dto.AuthorizePaymentResponse{
		TransactionID: data.TransactionID.String(),
		Status:        data.Status,
		CreatedAt:     data.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *PaymentHandler) logErr(c *gin.Context, err error) {
	h.log.Err(err).Str("path", c.FullPath()).Msg(err.Error())
}

func mapPaymentErrStatusCode(err error) int {
	switch {
	case errors.Is(err, domain.ErrTokenizeCard):
		return http.StatusNotAcceptable
	case errors.Is(err, domain.ErrDuplicateIdempotencyKey):
		return http.StatusConflict
	case errors.Is(err, domain.ErrGatewayRejected):
		return http.StatusPaymentRequired
	default:
		return http.StatusInternalServerError
	}
}
