package router

import (
	"card-payment-service/internal/gateway"
	"card-payment-service/internal/handler"
	"card-payment-service/internal/logger"
	"card-payment-service/internal/middleware"
	"card-payment-service/internal/repository"
	"card-payment-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB, redis *redis.Client) {
	// versioning
	v1 := r.Group("/v1")

	registerMerchant(v1, db)
	registerPayment(v1, db)
}

func registerMerchant(rg *gin.RouterGroup, db *gorm.DB) {
	merchantRepo := repository.NewMerchantRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)

	logger := logger.With("merchant_service")
	merchantService := service.NewMerchantService(merchantRepo, apiKeyRepo, logger)

	merchantHandler := handler.NewMerchantHandler(merchantService, logger)

	merchant := rg.Group("/merchants")
	{
		merchant.POST("/register", merchantHandler.Register)
		merchant.PATCH("/activate", merchantHandler.Activate)
	}
}

func registerPayment(rg *gin.RouterGroup, db *gorm.DB) {
	logger := logger.With("payment_service")

	txRepo := repository.NewTransactionRepository(db)
	idemRepo := repository.NewIdempotencyKeyRepository(db)
	gateway := gateway.NewMockGateway()

	paymentService := service.NewPaymentService(txRepo, idemRepo, gateway, logger)

	paymentHandler := handler.NewPaymentHandler(paymentService, logger)

	payment := rg.Group("/payments")
	{
		payment.POST("/authorize", middleware.MerchantAuth(), middleware.RequireIdempotencyKey(), paymentHandler.Authorize)
		payment.POST("/:transaction_id/capture", middleware.MerchantAuth(), paymentHandler.Capture)
	}
}
