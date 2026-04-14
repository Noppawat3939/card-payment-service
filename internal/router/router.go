package router

import (
	"card-payment-service/internal/handler"
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

	merchantService := service.NewMerchantService(merchantRepo, apiKeyRepo)

	merchantHandler := handler.NewMerchantHandler(merchantService)

	merchant := rg.Group("/merchants")
	{
		merchant.POST("/register", merchantHandler.Register)
	}
}
