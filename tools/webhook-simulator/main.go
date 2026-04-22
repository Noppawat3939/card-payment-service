package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// initialize configs
	_ = godotenv.Load(".env.local")

	targetURL := os.Getenv("SIMULATOR_TARGET_URL")
	secret := os.Getenv("SIMULATOR_WEBHOOK_SECRET")
	port := os.Getenv("SIMULATOR_PORT")

	if secret == "" {
		log.Fatal("SIMULATOR_WEBHOOK_SECRET is required")
	}
	if port == "" {
		port = "8081"
	}

	sim := NewSimulator(targetURL, secret)

	// setup router
	gin.SetMode(gin.DebugMode)
	r := gin.New()
	r.Use(gin.Logger())

	RegisterWebhookRoutes(r, sim)

	// run simulator server
	if err := r.Run(":" + port); err != nil {
		log.Fatal("failed to start simulator:", err)
	}
	log.Print("running simulator")
}
