package middleware

import (
	"card-payment-service/internal/domain"
	"card-payment-service/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const MerchantIDKey = "merchant_id"

func MerchantAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-Merchant-ID")
		if key == "" {
			response.Unauthorized(c, domain.ErrMissingMerchantID.Error())
			c.Abort()
			return
		}

		parsed, err := uuid.Parse(key)
		if err != nil {
			response.Unauthorized(c, "invalid merchant_id format")
			c.Abort()
			return
		}

		c.Set(MerchantIDKey, parsed)
		c.Next()
	}
}
