package middleware

import (
	"card-payment-service/internal/domain"
	"card-payment-service/internal/response"
	"card-payment-service/internal/service/auth"

	"github.com/gin-gonic/gin"
)

const MerchantIDKey = "merchant_id"

func MerchantAuth(as auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-KEY")
		if apiKey == "" {
			response.Unauthorized(c, domain.ErrMissingApiKey.Error())
			c.Abort()
			return
		}

		merchant, err := as.ValidateAPIKey(c, apiKey)
		if err != nil {
			response.Unauthorized(c, domain.ErrInvalidApiKey.Error())
			c.Abort()
			return
		}

		c.Set(MerchantIDKey, merchant.ID)
		c.Next()
	}
}
