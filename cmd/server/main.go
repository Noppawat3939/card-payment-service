package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	log.Println("server started on :8080")

	err := router.Run(":8080")
	if err != nil {
		log.Fatal("failed to start server:", err)
	}
}
