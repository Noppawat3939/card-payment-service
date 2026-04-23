package handler

import (
	"card-payment-service/internal/handler/dto"
	"card-payment-service/internal/handler/mapper"
	"card-payment-service/internal/httpx"
	"card-payment-service/internal/response"
	"card-payment-service/internal/service/payment"
	"net/http"

	"github.com/gin-gonic/gin"
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

	input := mapper.ToChargeInput(c, req)

	authorizedResp, err := h.paymentService.Authorize(c, input)
	if err != nil {
		response.FromError(c, err)
		return
	}

	data := dto.AuthorizePaymentResponse{
		TransactionID: authorizedResp.TransactionID,
		Status:        authorizedResp.Status,
	}
	response.Created(c, data)
}

func (h *PaymentHandler) Capture(c *gin.Context) {
	meta, err := httpx.GetPaymentMeta(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	capturedResp, err := h.paymentService.Capture(c, payment.CaptureInput{
		TransactionID: meta.TransactionID,
		MerchantID:    meta.MerchantID,
	})
	if err != nil {
		response.FromError(c, err)
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

	input := mapper.ToChargeInput(c, req)

	chargedResp, err := h.paymentService.Charge(c, input)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c, &dto.CapturePaymentResponse{
		TransactionID: chargedResp.TransactionID,
		Status:        chargedResp.Status,
	})
}

func (h *PaymentHandler) Void(c *gin.Context) {
	meta, err := httpx.GetPaymentMeta(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	voidedResp, err := h.paymentService.Void(c, payment.VoidInput{
		TransactionID: meta.TransactionID,
		MerchantID:    meta.MerchantID,
	})
	if err != nil {
		response.FromError(c, err)
		return
	}

	data := dto.VoidPaymentResponse{
		TransactionID: voidedResp.TransactionID,
		Status:        voidedResp.Status,
	}
	response.OK(c, data)
}

func (h *PaymentHandler) Refund(c *gin.Context) {
	meta, err := httpx.GetPaymentMeta(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	refundedResp, err := h.paymentService.Refund(c, payment.RefundInput{
		TransactionID:  meta.TransactionID,
		MerchantID:     meta.MerchantID,
		IdempotencyKey: meta.IdempotencyKey,
	})
	if err != nil {
		response.FromError(c, err)
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
