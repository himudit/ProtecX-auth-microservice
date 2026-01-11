package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"authService/config"
	"authService/internal/controllers"
	ratelimiter "authService/internal/middlewares"
	"authService/internal/models"
	"authService/internal/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load environment variables from .env only in development
	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			log.Println("⚠️ No .env file found, using system environment")
		}
	}

	config.ConnectRedis()
	config.ConnectMongo()
	models.InitCollections()
	config.LoadRSAKeys()

	r := gin.Default()

	r.Use(ratelimiter.RateLimiter(config.RDB))

	authController := controllers.NewAuthController(config.RDB)

	routes.AuthRoutes(r, authController)

	log.Println("🚀 Server running on :8080")
	r.Run(":8080")
}
