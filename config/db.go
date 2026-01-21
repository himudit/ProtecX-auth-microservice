package config

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB holds the PostgreSQL connection pool
type DB struct {
	Pool *pgxpool.Pool
}

// New creates and returns a new PostgreSQL connection pool
func New(ctx context.Context) *DB {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		log.Fatalf("failed to parse DATABASE_URL: %v", err)
	}

	// 🔧 Neon / serverless friendly defaults
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("failed to create db pool: %v", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	log.Println("✅ PostgreSQL connected (authService)")

	return &DB{Pool: pool}
}

// Close closes the database connection pool
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
		log.Println("🛑 PostgreSQL connection closed")
	}
}
