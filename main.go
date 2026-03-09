package main

import (
	"context"
	"log"
	"net"
	"os"

	"github.com/joho/godotenv"

	"authService/config"

	"google.golang.org/grpc"

	handlers "authService/internal/grpc"
	"authService/internal/grpc/interceptors"
	"authService/internal/repositories/postgres"
	"authService/internal/services"

	authv1 "authService/proto/proto/iam/v1"
)

func main() {

	ctx := context.Background()

	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			log.Println("⚠️ No .env file found, using system environment")
		}
	}

	// infra
	config.ConnectRedis()
	db := config.New(ctx)
	config.LoadRSAKeys()

	// repositories
	projectUserRepo := postgres.NewProjectUserRepository(db.Pool)
	projectJWTRepo := postgres.NewProjectJwtKeyRepository(db.Pool)

	// services
	authService := services.NewAuthService(projectUserRepo, projectJWTRepo, config.RDB)

	// handler
	authHandler := handlers.NewAuthHandler(authService)

	// grpc server
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.ProjectContextInterceptor(),
			interceptors.RateLimiterInterceptor(config.RDB),
			interceptors.AuthInterceptor(projectJWTRepo),
		),
	)

	// register service
	authv1.RegisterAuthServiceServer(grpcServer, authHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("❌ Failed to listen: %v", err)
	}

	log.Printf("🚀 gRPC server running on :%s", port)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("❌ Failed to serve: %v", err)
	}
}
