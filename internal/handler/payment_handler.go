package handler

import (
	"card-payment-service/internal/domain"
	"card-payment-service/internal/handler/dto"
	"card-payment-service/internal/middleware"
	"card-payment-service/internal/response"
	"card-payment-service/internal/service"
	"errors"
	"net/http"
	"time"

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

	input := buildAuthorizeInput(c, req, domain.AuthorizeCapture)

	data, err := h.paymentService.Authorize(c, *input)
	if err != nil {
		statusCode := mapPaymentErrStatusCode(err)
		response.Error(c, statusCode, err.Error())
		return
	}

	response.Created(c, dto.AuthorizePaymentResponse{
		TransactionID: data.TransactionID.String(),
		Status:        data.Status,
		CreatedAt:     formatTime(data.CreatedAt),
	})
}

func (h *PaymentHandler) Capture(c *gin.Context) {
	transactionIDStr := c.Param("transaction_id")
	if transactionIDStr == "" {
		h.logErr(c, errors.New("missing transaction_id"))
		response.BadRequest(c)
		return
	}
	transactionID, err := uuid.Parse(transactionIDStr)
	if err != nil {
		h.logErr(c, err)
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	merchantID := c.MustGet(middleware.MerchantIDKey).(uuid.UUID)

	data, err := h.paymentService.Capture(c, service.CaptureInput{
		TransactionID: transactionID,
		MerchantID:    merchantID,
	})
	if err != nil {
		statusCode := mapPaymentErrStatusCode(err)
		response.Error(c, statusCode, err.Error())
		return
	}

	response.OK(c, &dto.CapturePaymentResponse{
		TransactionID: data.TransactionID.String(),
		Status:        data.Status,
		CapturedAt:    formatTime(data.CapturedAt),
	})
}

func (h *PaymentHandler) Charge(c *gin.Context) {
	var req dto.AuthorizePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logErr(c, err)
		response.BadRequest(c)
		return
	}

	// process authorize
	input := buildAuthorizeInput(c, req, domain.DirectCharge)
	authorizeResp, err := h.paymentService.Authorize(c, *input)
	if err != nil {
		statusCode := mapPaymentErrStatusCode(err)
		response.Error(c, statusCode, err.Error())
		return
	}

	// process capture
	captureResp, err := h.paymentService.Capture(c, service.CaptureInput{
		TransactionID: authorizeResp.TransactionID,
		MerchantID:    input.MerchantID,
	})
	if err != nil {
		statusCode := mapPaymentErrStatusCode(err)
		response.Error(c, statusCode, err.Error())
		return
	}

	response.OK(c, &dto.CapturePaymentResponse{
		TransactionID: captureResp.TransactionID.String(),
		Status:        captureResp.Status,
		CapturedAt:    formatTime(captureResp.CapturedAt),
	})
}

func (h *PaymentHandler) logErr(c *gin.Context, err error) {
	h.log.Err(err).Str("path", c.FullPath()).Msg(err.Error())
}

func mapPaymentErrStatusCode(err error) int {
	switch {
	case errors.Is(err, domain.ErrTokenizeCard):
		return http.StatusNotAcceptable
	case errors.Is(err, domain.ErrDuplicateIdempotencyKey),
		errors.Is(err, domain.ErrDuplicateRequest):
		return http.StatusConflict
	case errors.Is(err, domain.ErrGatewayRejected):
		return http.StatusPaymentRequired
	case errors.Is(err, domain.ErrCardInforInvalid),
		errors.Is(err, domain.ErrCardAmoutInvalid),
		errors.Is(err, domain.ErrCardCaptureFailed),
		errors.Is(err, domain.ErrCardDeclinded),
		errors.Is(err, domain.ErrExpiredCard),
		errors.Is(err, domain.ErrInsufficientFunds):
		return http.StatusPaymentRequired

	default:
		return http.StatusInternalServerError
	}
}

func buildAuthorizeInput(c *gin.Context, req dto.AuthorizePaymentRequest, paymentType domain.PaymentType) *service.AuthorizeInput {
	merchantID := c.MustGet(middleware.MerchantIDKey).(uuid.UUID)
	idemKey := c.MustGet(middleware.IdempotencyKeyContextKey).(string)

	return &service.AuthorizeInput{
		Amount:         req.Amount,
		CardNumber:     req.CardNumber,
		Currency:       req.Currency,
		CVV:            req.CVV,
		Description:    req.Description,
		ExpiryMonth:    req.ExpiryMonth,
		ExpiryYear:     req.ExpiryYear,
		IdempotencyKey: idemKey,
		MerchantID:     merchantID,
		PaymentType:    paymentType,
	}
}

func formatTime(t time.Time) string {
	return t.Format("2006-01-02T15:04:05Z07:00")
}
