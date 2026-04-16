package middleware

import (
	"card-payment-service/internal/response"

	"github.com/gin-gonic/gin"
)

const IdempotencyKeyContextKey = "idempotency_key"

func RequireIdempotencyKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("Idempotency-Key")
		if key == "" {
			response.Unauthorized(c, "missing idempotency key")
			c.Abort()
			return
		}

		c.Set(IdempotencyKeyContextKey, key)
		c.Next()
	}
}
