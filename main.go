package main

import (
	"log"

	"authService/config"
	"authService/internal/controllers"
	ratelimiter "authService/internal/middlewares"
	"authService/internal/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Connect Redis
	config.ConnectRedis()

	// Create Gin router
	r := gin.Default()

	// Global rate limiter middleware
	r.Use(ratelimiter.RateLimiter(config.RDB))

	// Initialize controller
	authController := controllers.NewAuthController()

	// Register routes
	routes.AuthRoutes(r, authController)

	log.Println("🚀 Server running on :8080")
	r.Run(":8080")
}
