package handler

import (
	"card-payment-service/internal/domain"
	"card-payment-service/internal/handler/dto"
	"card-payment-service/internal/response"
	"card-payment-service/internal/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MerchantHandler struct {
	merchantService *service.MerchantService
}

func NewMerchantHandler(merchantService *service.MerchantService) *MerchantHandler {
	return &MerchantHandler{merchantService}
}

func (h *MerchantHandler) Register(c *gin.Context) {
	var req dto.RegisterMerchantRequest

	if e := c.ShouldBindJSON(&req); e != nil {
		response.Error(c, http.StatusBadRequest, domain.ErrBodyInvalid.Error())
		return
	}

	data, e := h.merchantService.Register(c, service.RegisterInput{
		Name:       req.Name,
		Email:      req.Name,
		WebhookURL: req.WebhookURL,
	})
	if e != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(e, domain.ErrMerchantAlreadyExists) {
			statusCode = http.StatusNotAcceptable
		}

		response.Error(c, statusCode, e.Error())
		return
	}

	response.Created(c, &dto.RegisterMerchantResponse{
		APIKey:     data.APIKey,
		APISecret:  data.APISecret,
		MerchantID: data.MerchantID,
		Status:     data.Status,
	})
}
