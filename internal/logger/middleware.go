package logger

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func GinLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		ctx.Next()

		log.Logger.Info().
			Str("method", ctx.Request.Method).
			Str("path", ctx.Request.URL.Path).
			Str("request_id", ctx.GetHeader("X-Request-ID")).
			Int("status", ctx.Writer.Status()).
			Dur("latency", time.Since(start)).
			Msg("request")

	}
}
