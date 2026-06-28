package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/mailgo/backend/internal/infrastructure/brevo"
	"github.com/mailgo/backend/internal/infrastructure/mailcow"
	"github.com/mailgo/backend/internal/infrastructure/postgres"
	"github.com/mailgo/backend/internal/worker"
)

func main() {
	// Load environment variables
	godotenv.Load()

	// Setup logger
	setupLogger()

	// Initialize database
	db, err := initDatabase()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	// Initialize Redis client for Asynq
	redisOpt := asynq.RedisClientOpt{
		Addr: getEnv("REDIS_ADDR", "localhost:6379"),
		DB:   0,
	}

	// Initialize infrastructure adapters
	mailServerAdapter := mailcow.NewAdapter(
		getEnv("MAILCOW_API_URL", "http://mailcow:8080/api/v1"),
		getEnv("MAILCOW_API_KEY", ""),
		30*time.Second,
	)

	emailSenderAdapter := brevo.NewAdapter(
		getEnv("BREVO_API_URL", "https://api.brevo.com/v3"),
		getEnv("BREVO_API_KEY", ""),
		getEnv("BREVO_FROM_EMAIL", "noreply@mailgo.com"),
		getEnv("BREVO_FROM_NAME", "MailGo"),
		30*time.Second,
	)

	// Initialize repositories
	mailboxRepo := postgres.NewMailboxRepository(db)
	domainRepo := postgres.NewDomainRepository(db)

	// Initialize task handlers
	handlers := worker.NewHandlers(
		mailboxRepo,
		domainRepo,
		mailServerAdapter,
		emailSenderAdapter,
		emailSenderAdapter,
	)

	// Create Asynq server
	srv := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,  // 60% priority
				"default":  3,  // 30% priority
				"low":      1,  // 10% priority
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Error().
					Err(err).
					Str("task", task.Type()).
					Bytes("payload", task.Payload()).
					Msg("Task failed")
			}),
			LogLevel: asynq.InfoLevel,
		},
	)

	// Register task handlers
	mux := asynq.NewServeMux()
	mux.HandleFunc(worker.TypeMailboxProvision, handlers.HandleMailboxProvision)
	mux.HandleFunc(worker.TypeMailboxDelete, handlers.HandleMailboxDelete)
	mux.HandleFunc(worker.TypeDomainSetup, handlers.HandleDomainSetup)
	mux.HandleFunc(worker.TypeDomainVerification, handlers.HandleDomainVerification)
	mux.HandleFunc(worker.TypeEmailSend, handlers.HandleEmailSend)
	mux.HandleFunc(worker.TypeAuditLogProcess, handlers.HandleAuditLogProcess)

	// Start worker
	log.Info().Msg("Starting Asynq worker...")
	
	if err := srv.Run(mux); err != nil {
		log.Fatal().Err(err).Msg("Failed to start worker")
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down worker...")
	srv.Shutdown()
	log.Info().Msg("Worker exited")
}

func setupLogger() {
	logLevel := getEnv("LOG_LEVEL", "info")
	
	switch logLevel {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
}

func initDatabase() (*pgxpool.Pool, error) {
	dbURL := getEnv("DATABASE_URL", "")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL not set")
	}

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	log.Info().Msg("Connected to PostgreSQL")
	return pool, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
