package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"

	"authService/config"
	"authService/internal/controllers"
	"authService/internal/repositories/postgres"
	"authService/internal/routes"
	"authService/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load environment variables from .env only in development
	ctx := context.Background()
	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			log.Println("⚠️ No .env file found, using system environment")
		}
	}

	config.ConnectRedis()
	db := config.New(ctx)
	// models.InitCollections()
	config.LoadRSAKeys()

	r := gin.Default()

	projectUserRepo := postgres.NewProjectUserRepository(db.Pool)

	authService := services.NewAuthService(projectUserRepo)

	authController := controllers.NewAuthController(config.RDB, authService)

	routes.AuthRoutes(r, authController)

	log.Println("🚀 Server running on :8080")
	r.Run(":8080")
}
