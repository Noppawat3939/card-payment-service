package router

import (
	"card-payment-service/internal/gateway"
	"card-payment-service/internal/handler"
	appRedis "card-payment-service/internal/infra/redis"
	"card-payment-service/internal/logger"
	"card-payment-service/internal/middleware"
	"card-payment-service/internal/repository"
	"card-payment-service/internal/service/auth"
	"card-payment-service/internal/service/merchant"
	"card-payment-service/internal/service/payment"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Config struct {
	rg     *gin.RouterGroup
	db     *gorm.DB
	client *redis.Client
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB, client *redis.Client) {
	// versioning
	v1 := r.Group("/v1")
	cfg := &Config{rg: v1, db: db, client: client}

	registerMerchant(cfg)
	registerPayment(cfg)
}

func registerMerchant(cfg *Config) {
	merchantRepo := repository.NewMerchantRepository(cfg.db)
	apiKeyRepo := repository.NewAPIKeyRepository(cfg.db)

	logger := logger.With("merchant_service")
	merchantService := merchant.NewMerchantService(merchantRepo, apiKeyRepo, logger)

	merchantHandler := handler.NewMerchantHandler(merchantService, logger)

	merchant := cfg.rg.Group("/merchants")
	{
		merchant.POST("/register", merchantHandler.Register)
		merchant.PATCH("/activate", merchantHandler.Activate)
	}
}

func registerPayment(cfg *Config) {
	logger := logger.With("payment_service")

	txRepo := repository.NewTransactionRepository(cfg.db)
	idemRepo := repository.NewIdempotencyKeyRepository(cfg.db)
	refundRepo := repository.NewRefundRepository(cfg.db)
	gateway := gateway.NewMockGateway()

	redisLocker := appRedis.NewRedisLocker(cfg.client)
	paymentService := payment.NewPaymentService(txRepo, idemRepo, refundRepo, gateway, redisLocker, logger)
	authService := initAuthService(cfg)

	paymentHandler := handler.NewPaymentHandler(paymentService, logger)

	payment := cfg.rg.Group("/payments")
	payment.Use(middleware.MerchantAuth(authService))

	// routes required idempotency key
	withIdemKey := payment.Group("")
	withIdemKey.Use(middleware.RequireIdempotencyKey())
	{
		withIdemKey.POST("/authorize", paymentHandler.Authorize)
		withIdemKey.POST("/charge", paymentHandler.Charge)
		withIdemKey.POST("/refund/:transaction_id", paymentHandler.Refund)
	}
	// routes without idempotency key
	payment.POST("/:transaction_id/capture", paymentHandler.Capture)
	payment.POST(":transaction_id/void", paymentHandler.Void)
}

func initAuthService(cfg *Config) auth.AuthService {
	apiRepo := repository.NewAPIKeyRepository(cfg.db)
	merchantRepo := repository.NewMerchantRepository(cfg.db)
	service := auth.NewAuthService(apiRepo, merchantRepo)

	return service
}
