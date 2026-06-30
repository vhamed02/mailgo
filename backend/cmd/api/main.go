package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/mailgo/backend/internal/api/rest/handlers"
	restmiddleware "github.com/mailgo/backend/internal/api/rest/middleware"
	"github.com/mailgo/backend/internal/application"
	"github.com/mailgo/backend/internal/infrastructure/brevo"
	imapinfra "github.com/mailgo/backend/internal/infrastructure/imap"
	"github.com/mailgo/backend/internal/infrastructure/mailcow"
	"github.com/mailgo/backend/internal/infrastructure/postgres"
	"github.com/mailgo/backend/internal/infrastructure/queue"
	rediscache "github.com/mailgo/backend/internal/infrastructure/redis"
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

	// Run migrations automatically on startup
	if err := runMigrations(db); err != nil {
		log.Fatal().Err(err).Msg("Failed to run database migrations")
	}

	// Initialize Redis
	redisClient := initRedis()
	defer redisClient.Close()

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
	_ = emailSenderAdapter

	cacheAdapter := rediscache.NewAdapter(redisClient)
	_ = cacheAdapter

	queueAdapter := queue.NewAdapter(redisClient)

	imapPort := 993
	imapTLS := true
	if p := getEnv("IMAP_PORT", ""); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			imapPort = v
		}
	}
	if getEnv("IMAP_TLS", "true") == "false" {
		imapTLS = false
	}
	imapAdapter := imapinfra.NewAdapter(getEnv("IMAP_HOST", "mailcow"), imapPort, imapTLS)

	mailService := application.NewMailService(imapAdapter, emailSenderAdapter, getEnv("BREVO_FROM_NAME", "MailGo"))

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	orgRepo := postgres.NewOrganizationRepository(db)
	orgUserRepo := postgres.NewOrganizationUserRepository(db)
	domainRepo := postgres.NewDomainRepository(db)
	mailboxRepo := postgres.NewMailboxRepository(db)
	aliasRepo := postgres.NewAliasRepository(db)
	_ = aliasRepo // Will be used for alias management

	quotaRepo := postgres.NewQuotaRepository(db)
	auditRepo := postgres.NewAuditLogRepository(db)
	txManager := postgres.NewTxManager(db)

	// Initialize services
	authService := application.NewAuthService(
		userRepo,
		orgRepo,
		orgUserRepo,
		quotaRepo,
		auditRepo,
		txManager,
		queueAdapter,
		getEnv("JWT_SECRET", "change-this-in-production"),
		15*time.Minute,
	)

	mailboxService := application.NewMailboxService(
		mailboxRepo,
		domainRepo,
		quotaRepo,
		auditRepo,
		mailServerAdapter,
		cacheAdapter,
		queueAdapter,
	)

	domainService := application.NewDomainService(
		domainRepo,
		quotaRepo,
		auditRepo,
		queueAdapter,
		getEnv("MAIL_SERVER_HOSTNAME", ""),
	)

	// Initialize Echo server
	e := echo.New()
	e.HideBanner = true
	e.Validator = &CustomValidator{validator: validator.New()}

	// Global middleware
	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(100)))

	// Health check endpoint (no auth required)
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "healthy",
			"service": "mailgo-api",
		})
	})

	// Metrics endpoint (no auth required)
	e.GET("/metrics", func(c echo.Context) error {
		// TODO: Implement Prometheus metrics
		return c.JSON(http.StatusOK, map[string]string{
			"status": "metrics endpoint",
		})
	})

	// API v1 routes
	v1 := e.Group("/api/v1")

	// Auth routes (no auth middleware)
	authHandler := handlers.NewAuthHandler(authService)
	v1.POST("/auth/register", authHandler.Register)
	v1.POST("/auth/login", authHandler.Login)
	v1.POST("/auth/logout", authHandler.Logout)

	// Protected routes (require authentication)
	protected := v1.Group("")
	protected.Use(restmiddleware.AuthMiddleware(authService))

	// User info
	protected.GET("/auth/me", authHandler.Me)

	// Mail routes authenticate via mailbox IMAP credentials
	// (X-Mailbox-Address / X-Mailbox-Password headers), not JWT.
	mailHandler := handlers.NewMailHandler(mailService)
	v1.GET("/mail/folders", mailHandler.ListFolders)
	v1.GET("/mail/folders/:folder/messages", mailHandler.ListMessages)
	v1.GET("/mail/messages/:uid", mailHandler.GetMessage)
	v1.PATCH("/mail/messages/:uid/read", mailHandler.MarkRead)
	v1.DELETE("/mail/messages/:uid", mailHandler.DeleteMessage)
	v1.POST("/mail/compose", mailHandler.Compose)
	v1.POST("/mail/reply", mailHandler.Reply)

	// Domain routes
	domainHandler := handlers.NewDomainHandler(domainService)
	protected.GET("/domains", domainHandler.List)
	protected.POST("/domains", domainHandler.Create)
	protected.GET("/domains/:id", domainHandler.Get)
	protected.DELETE("/domains/:id", domainHandler.Delete)
	protected.POST("/domains/:id/verify", domainHandler.Verify)
	protected.POST("/domains/:id/regenerate", domainHandler.Regenerate)

	// Mailbox routes
	mailboxHandler := handlers.NewMailboxHandler(mailboxService)
	protected.GET("/mailboxes", mailboxHandler.List)
	protected.POST("/mailboxes", mailboxHandler.Create)
	protected.GET("/mailboxes/:id", mailboxHandler.Get)
	protected.PATCH("/mailboxes/:id", mailboxHandler.Update)
	protected.DELETE("/mailboxes/:id", mailboxHandler.Delete)
	protected.POST("/mailboxes/:id/suspend", mailboxHandler.Suspend)
	protected.POST("/mailboxes/:id/unsuspend", mailboxHandler.Unsuspend)

	// Start server
	port := getEnv("APP_PORT", "8080")
	go func() {
		log.Info().Msgf("Starting API server on port %s", port)
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
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

func initRedis() *redis.Client {
	redisURL := getEnv("REDIS_URL", "redis://localhost:6379/0")

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse Redis URL")
	}

	client := redis.NewClient(opt)

	// Test connection
	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}

	log.Info().Msg("Connected to Redis")
	return client
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// CustomValidator wraps the go-playground validator
type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return err
	}
	return nil
}

// runMigrations executes all .up.sql migration files that haven't been applied yet.
// It uses a simple schema_migrations table to track applied migrations.
func runMigrations(db *pgxpool.Pool) error {
	ctx := context.Background()

	// Ensure migrations tracking table exists
	_, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// If the main schema already exists but schema_migrations is empty (e.g. first
	// run after adding this runner to an existing DB), record it as applied so we
	// don't try to re-execute it.
	var orgExists bool
	_ = db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'organizations'
		)
	`).Scan(&orgExists)
	if orgExists {
		_, _ = db.Exec(ctx, `
			INSERT INTO schema_migrations (version) VALUES ('000001_initial_schema')
			ON CONFLICT DO NOTHING
		`)
	}

	// Migration files in order
	migrations := []struct {
		version string
		path    string
	}{
		{"000001_initial_schema", "migrations/000001_initial_schema.up.sql"},
		{"000002_domain_email_setup", "migrations/000002_domain_email_setup.up.sql"},
		{"000003_domain_dns_records", "migrations/000003_domain_dns_records.up.sql"},
		{"000004_domain_mx_record", "migrations/000004_domain_mx_record.up.sql"},
	}

	for _, m := range migrations {
		// Check if already applied
		var count int
		err := db.QueryRow(ctx,
			`SELECT COUNT(*) FROM schema_migrations WHERE version = $1`, m.version,
		).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration %s: %w", m.version, err)
		}
		if count > 0 {
			log.Info().Str("version", m.version).Msg("Migration already applied, skipping")
			continue
		}

		// Read and execute migration file
		sql, err := os.ReadFile(m.path)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", m.path, err)
		}

		if _, err := db.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", m.version, err)
		}

		// Record as applied
		if _, err := db.Exec(ctx,
			`INSERT INTO schema_migrations (version) VALUES ($1)`, m.version,
		); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", m.version, err)
		}

		log.Info().Str("version", m.version).Msg("Migration applied successfully")
	}

	return nil
}
