package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JullMol/titan-ledger/config"
	"github.com/JullMol/titan-ledger/internal/adapters/handler/http"
	"github.com/JullMol/titan-ledger/internal/adapters/repository/postgres"
	"github.com/JullMol/titan-ledger/internal/core/services"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Use DATABASE_URL if available (Railway standard), otherwise build from parts
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode,
		)
	}
	log.Printf("Connecting to database...")

	ctx := context.Background()
	dbPool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		log.Fatalf("Database ping failed: %v\n", err)
	}
	log.Println("Database Connected!")

	// Auto-migrate: create tables if not exist
	if err := runMigrations(ctx, dbPool); err != nil {
		log.Fatalf("Migration failed: %v\n", err)
	}
	log.Println("Migrations completed!")

	walletRepo := postgres.NewWalletRepository(dbPool)
	txRepo := postgres.NewTransactionRepository(dbPool)
	txManager := postgres.NewPgTxManager(dbPool)

	wallletService := services.NewWalletService(walletRepo)
	transferService := services.NewTransferService(walletRepo, txRepo, txManager)
	
	handler := http.NewTitanHandler(wallletService, transferService)

	app := fiber.New(fiber.Config{
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})
	app.Use(logger.New())

	api := app.Group("/api/v1")
	api.Post("/wallets", handler.CreateWallet)
	api.Get("/wallets/:id", handler.GetBalance)
	api.Post("/transfer", handler.Transfer)

	// Health check endpoint for container orchestration
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Serve static files for API playground
	app.Static("/", "./static")

	go func() {
		// Use PORT from env (Railway standard), fallback to config
		port := os.Getenv("PORT")
		if port == "" {
			port = cfg.ServerPort
		}
		log.Printf("Server running on port %s", port)
		if err := app.Listen(":" + port); err != nil {
			log.Panic(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	log.Println("\nShutdown signal received")

	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("Shutting down Fiber server")
	if err := app.Shutdown(); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server shutdown gracefully!")
}

// runMigrations creates database tables if they don't exist
func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrations := `
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

		CREATE TABLE IF NOT EXISTS wallets (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id VARCHAR(255) NOT NULL,
			balance BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0),
			currency VARCHAR(3) NOT NULL DEFAULT 'IDR',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS transactions (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			wallet_id UUID NOT NULL REFERENCES wallets(id),
			amount BIGINT NOT NULL,
			reference_id VARCHAR(255) NOT NULL,
			type VARCHAR(20) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_transactions_wallet_id ON transactions(wallet_id);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_reference_unique ON transactions(reference_id);
	`
	_, err := pool.Exec(ctx, migrations)
	return err
}