package main

import (
	"card-payment-service/internal/config"
	"card-payment-service/internal/infra/database"
	ridisClient "card-payment-service/internal/infra/redis"
	"card-payment-service/internal/logger"
	"card-payment-service/internal/router"
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func main() {
	// load configs
	cfg := config.Load()

	// init logger
	logger.Init(logger.Config{
		Level:  "Info",
		Pretty: os.Getenv("APP_ENV") == "develop",
	})

	log.Info().Str("env", os.Getenv("APP_ENV")).Msg("starting server")

	// connect to database
	db, err := database.NewPostgres(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect database")
	}

	defer database.CloseDB(db)

	// connect to redis
	client, err := ridisClient.NewRedis(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect redis")
	}
	defer client.Close()

	// initalize router
	r := gin.Default()
	r.Use(logger.GinLogger())
	r.Use(gin.Recovery())
	router.RegisterRoutes(r, db, client)

	// create http server
	srv := startServer(cfg.AppPort, r)

	gracefulShutdown(srv, db, client)
}

func startServer(port string, router *gin.Engine) *http.Server {
	srv := http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Info().Str("port", port).Msg("server listening")
		if e := srv.ListenAndServe(); e != nil && e != http.ErrServerClosed {
			log.Fatal().Err(e).Msg("server failed to start")
		}
	}()

	return &srv
}

func gracefulShutdown(
	srv *http.Server,
	db *gorm.DB,
	redisClient *redis.Client,
) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit

	log.Info().
		Str("signal", sig.String()).
		Msg("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("server forced to shutdown")
	} else {
		log.Info().Msg("server shutdown complete")
	}

	if err := redisClient.Close(); err != nil {
		log.Error().Err(err).Msg("failed to close redis")
	} else {
		log.Info().Msg("redis connection closed")
	}

	database.CloseDB(db)
	log.Info().Msg("database connection closed")

	log.Info().Msg("application exited")
}
