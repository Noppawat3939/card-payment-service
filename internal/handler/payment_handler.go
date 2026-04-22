package handler

import (
	"card-payment-service/internal/domain"
	"card-payment-service/internal/handler/dto"
	"card-payment-service/internal/middleware"
	"card-payment-service/internal/response"
	"card-payment-service/internal/service/payment"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type PaymentHandler struct {
	paymentService *payment.PaymentService
	log            zerolog.Logger
}

func NewPaymentHandler(
	paymentService *payment.PaymentService,
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
		TransactionID: authorizedResp.TransactionID,
		Status:        authorizedResp.Status,
	}
	response.Created(c, data)
}

func (h *PaymentHandler) Capture(c *gin.Context) {
	transactionID := h.getTransactionIDParam(c)

	merchantID := h.getMerchantIDHeader(c)

	capturedResp, err := h.paymentService.Capture(c, payment.CaptureInput{
		TransactionID: *transactionID,
		MerchantID:    *merchantID,
	})
	if err != nil {
		status := mapPaymentErrStatusCode(err)
		response.Error(c, status, err.Error())
		return
	}

	data := dto.CapturePaymentResponse{
		TransactionID: capturedResp.TransactionID,
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
		TransactionID: chargedResp.TransactionID,
		Status:        chargedResp.Status,
	})
}

func (h *PaymentHandler) Void(c *gin.Context) {
	transactionID := h.getTransactionIDParam(c)
	merchantID := h.getMerchantIDHeader(c)

	voidedResp, err := h.paymentService.Void(c, payment.VoidInput{
		TransactionID: *transactionID,
		MerchantID:    *merchantID,
	})
	if err != nil {
		status := mapPaymentErrStatusCode(err)
		response.Error(c, status, err.Error())
		return
	}

	data := dto.VoidPaymentResponse{
		TransactionID: voidedResp.TransactionID,
		Status:        voidedResp.Status,
	}
	response.OK(c, data)
}

func (h *PaymentHandler) Refund(c *gin.Context) {
	transactionID := h.getTransactionIDParam(c)
	merchantID := h.getMerchantIDHeader(c)
	idemKey := c.MustGet(middleware.IdempotencyKeyContextKey).(uuid.UUID)

	refundedResp, err := h.paymentService.Refund(c, payment.RefundInput{
		TransactionID:  *transactionID,
		MerchantID:     *merchantID,
		IdempotencyKey: idemKey,
	})
	if err != nil {
		status := mapPaymentErrStatusCode(err)
		response.Error(c, status, err.Error())
		return
	}

	data := dto.RefundResponse{
		RefundID: refundedResp.RefundID,
		Status:   refundedResp.Status,
	}
	response.OK(c, data)
}

func (h *PaymentHandler) logErr(c *gin.Context, err error) {
	h.log.Err(err).Str("path", c.FullPath()).Msg(err.Error())
}

func mapPaymentErrStatusCode(err error) int {
	switch {
	// 406
	case errors.Is(err, domain.ErrTokenizeCard):
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

func buildChargeInput(c *gin.Context, req dto.AuthorizePaymentRequest) *payment.ChargeInput {
	merchantID := c.MustGet(middleware.MerchantIDKey).(uuid.UUID)
	idemKey := c.MustGet(middleware.IdempotencyKeyContextKey).(string)

	return &payment.ChargeInput{
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

func (h *PaymentHandler) getTransactionIDParam(c *gin.Context) *uuid.UUID {
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

func (h *PaymentHandler) getMerchantIDHeader(c *gin.Context) *uuid.UUID {
	merchantID := c.MustGet(middleware.MerchantIDKey).(uuid.UUID)
	err := uuid.Validate(merchantID.String())
	if err != nil {
		h.logErr(c, errors.New("merchant_id invalid"))
		response.Unauthorized(c, domain.ErrMissingMerchantID.Error())
		return nil
	}

	return &merchantID
}
