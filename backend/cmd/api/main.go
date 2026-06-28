package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
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
	_ = emailSenderAdapter // Used by workers, not API yet

	cacheAdapter := rediscache.NewAdapter(redisClient)
	_ = cacheAdapter // Will be used for caching

	queueAdapter := queue.NewAdapter(redisClient)

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

	// Initialize services
	authService := application.NewAuthService(
		userRepo,
		orgRepo,
		orgUserRepo,
		quotaRepo,
		auditRepo,
		getEnv("JWT_SECRET", "change-this-in-production"),
		15*time.Minute,
	)

	mailboxService := application.NewMailboxService(
		mailboxRepo,
		domainRepo,
		quotaRepo,
		auditRepo,
		mailServerAdapter,
		queueAdapter,
	)
	_ = mailboxService // Will be used for mailbox handlers

	domainService := application.NewDomainService(
		domainRepo,
		quotaRepo,
		auditRepo,
		queueAdapter,
	)
	_ = domainService // Will be used for domain handlers

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
			"status": "healthy",
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

	// Organization routes
	// orgHandler := handlers.NewOrganizationHandler(orgService)
	// protected.GET("/organizations", orgHandler.List)
	// protected.GET("/organizations/:id", orgHandler.Get)

	// Domain routes
	// domainHandler := handlers.NewDomainHandler(domainService)
	// protected.GET("/domains", domainHandler.List)
	// protected.POST("/domains", domainHandler.Create)
	// protected.GET("/domains/:id", domainHandler.Get)
	// protected.DELETE("/domains/:id", domainHandler.Delete)
	// protected.POST("/domains/:id/verify", domainHandler.Verify)

	// Mailbox routes
	// mailboxHandler := handlers.NewMailboxHandler(mailboxService)
	// protected.GET("/mailboxes", mailboxHandler.List)
	// protected.POST("/mailboxes", mailboxHandler.Create)
	// protected.GET("/mailboxes/:id", mailboxHandler.Get)
	// protected.PATCH("/mailboxes/:id", mailboxHandler.Update)
	// protected.DELETE("/mailboxes/:id", mailboxHandler.Delete)
	// protected.POST("/mailboxes/:id/suspend", mailboxHandler.Suspend)
	// protected.POST("/mailboxes/:id/unsuspend", mailboxHandler.Unsuspend)

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
