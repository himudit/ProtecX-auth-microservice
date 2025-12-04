package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	Env       string
	DBUrl     string
	JwtSecret string
}

var App Config

func LoadConfig() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	App.Port = getEnv("PORT", "8080")
	App.Env = getEnv("ENV", "development")
	App.DBUrl = getEnv("DB_URL", "")
	App.JwtSecret = getEnv("JWT_SECRET", "")
	if App.JwtSecret == "" {
		log.Println("WARNING: JWT_SECRET is empty!")
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
