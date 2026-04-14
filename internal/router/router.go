package router

import (
	"card-payment-service/internal/handler"
	"card-payment-service/internal/logger"
	"card-payment-service/internal/repository"
	"card-payment-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB, redis *redis.Client) {
	// versioning
	version := r.Group("/v1")

	registerMerchant(version, db)
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
