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

func NewPaymentHandler(
	paymentService *service.PaymentService,
	log zerolog.Logger,
) *PaymentHandler {
	return &PaymentHandler{
		paymentService,
		log,
	}
}

func (h *PaymentHandler) Authorize(c *gin.Context) {
	var req dto.AuthorizePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logErr(c, err)
		response.BadRequest(c)
		return
	}

	input := buildChargeInput(c, req)

	authorizedResp, err := h.paymentService.Authorize(c, *input)
	if err != nil {
		status := mapPaymentErrStatusCode(err)
		response.Error(c, status, err.Error())
		return
	}

	data := dto.AuthorizePaymentResponse{
		TransactionID: authorizedResp.TransactionID.String(),
		Status:        authorizedResp.Status,
	}
	response.Created(c, data)
}

func (h *PaymentHandler) Capture(c *gin.Context) {
	transactionID := h.validateTransactionID(c)

	merchantID := c.MustGet(middleware.MerchantIDKey).(uuid.UUID)

	capturedResp, err := h.paymentService.Capture(c, service.CaptureInput{
		TransactionID: *transactionID,
		MerchantID:    merchantID,
	})
	if err != nil {
		status := mapPaymentErrStatusCode(err)
		response.Error(c, status, err.Error())
		return
	}

	data := dto.CapturePaymentResponse{
		TransactionID: capturedResp.TransactionID.String(),
		Status:        capturedResp.Status,
	}
	response.OK(c, data)
}

func (h *PaymentHandler) Charge(c *gin.Context) {
	var req dto.AuthorizePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logErr(c, err)
		response.BadRequest(c)
		return
	}

	input := buildChargeInput(c, req)

	chargedResp, err := h.paymentService.Charge(c, *input)
	if err != nil {
		statusCode := mapPaymentErrStatusCode(err)
		response.Error(c, statusCode, err.Error())
		return
	}

	response.OK(c, &dto.CapturePaymentResponse{
		TransactionID: chargedResp.TransactionID.String(),
		Status:        chargedResp.Status,
	})
}

func (h *PaymentHandler) Void(c *gin.Context) {
	var req dto.VoidPaymentResponse

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logErr(c, err)
		response.BadRequest(c)
		return
	}

	transactionID := h.validateTransactionID(c)
	merchantID := c.MustGet(middleware.MerchantIDKey).(uuid.UUID)

	voidedResp, err := h.paymentService.Void(c, service.VoidInput{TransactionID: *transactionID, MerchantID: merchantID})
	if err != nil {
		status := mapPaymentErrStatusCode(err)
		response.Error(c, status, err.Error())
		return
	}

	data := dto.VoidPaymentResponse{
		TransactionID: voidedResp.TransactionID.String(),
		Status:        voidedResp.Status,
	}
	response.OK(c, data)
}

func (h *PaymentHandler) logErr(c *gin.Context, err error) {
	h.log.Err(err).Str("path", c.FullPath()).Msg(err.Error())
}

func mapPaymentErrStatusCode(err error) int {
	switch {
	case errors.Is(err, domain.ErrTokenizeCard):
		return http.StatusNotAcceptable

	case errors.Is(err, domain.ErrDuplicateIdempotencyKey),
		errors.Is(err, domain.ErrTransactionAlreadyVoided),
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

func buildChargeInput(c *gin.Context, req dto.AuthorizePaymentRequest) *service.ChargeInput {
	merchantID := c.MustGet(middleware.MerchantIDKey).(uuid.UUID)
	idemKey := c.MustGet(middleware.IdempotencyKeyContextKey).(string)

	return &service.ChargeInput{
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

func (h *PaymentHandler) validateTransactionID(c *gin.Context) *uuid.UUID {
	txStr := c.Param("transaction_id")
	if txStr == "" {
		h.logErr(c, errors.New("body invalid missing transaction_id"))
		response.BadRequest(c)
		return nil
	}

	// parse uuid
	txID, err := uuid.Parse(txStr)

	if err != nil {
		h.logErr(c, err)
		response.Error(c, http.StatusBadRequest, err.Error())
		return nil
	}
	return &txID
}
