package handler

import (
	"card-payment-service/internal/handler/dto"
	"card-payment-service/internal/response"
	"card-payment-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type MerchantHandler struct {
	merchantService *service.MerchantService
	log             zerolog.Logger
}

func NewMerchantHandler(merchantService *service.MerchantService, log zerolog.Logger) *MerchantHandler {
	return &MerchantHandler{merchantService, log}
}

func (h *MerchantHandler) Register(c *gin.Context) {
	var req dto.RegisterMerchantRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logErr(c, err)
		response.BadRequest(c)
		return
	}

	data, err := h.merchantService.Register(c, service.RegisterInput{
		Name:       req.Name,
		Email:      req.Email,
		WebhookURL: req.WebhookURL,
	})
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Created(c, &dto.RegisterMerchantResponse{
		APIKey:     data.APIKey,
		MerchantID: data.MerchantID,
		Status:     data.Status,
	})
}

func (h *MerchantHandler) Activate(c *gin.Context) {
	var req dto.UpdateMerchantStatusRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logErr(c, err)
		response.BadRequest(c)
		return
	}

	merchant, err := h.merchantService.Activate(c, req.Email)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c, &dto.UpdateMerchantStatusResponse{
		Name:   merchant.Name,
		Email:  merchant.Email,
		Status: string(merchant.Status),
	})
}

func (h *MerchantHandler) logErr(c *gin.Context, err error) {
	h.log.Err(err).Str("path", c.FullPath()).Msg(err.Error())
}
