package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	// "github.com/mackb/releaseradar/internal/adapter/cache" // Удален неиспользуемый импорт
	"github.com/mackb/releaseradar/internal/adapter/github"
	"github.com/mackb/releaseradar/internal/adapter/persistence"
	"github.com/mackb/releaseradar/internal/adapter/telegram"
	"github.com/mackb/releaseradar/internal/usecase"
	"github.com/mackb/releaseradar/pkg/idempotency"
	"github.com/mackb/releaseradar/pkg/logger"
	"github.com/mackb/releaseradar/pkg/otel"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	// For Swagger documentation
	// _ "github.com/mackb/releaseradar/docs"
)

func init() {
	vipHook := viper.New()
	vipHook.AutomaticEnv()
	vipHook.SetEnvPrefix("RR")
	vipHook.SetDefault("LOG_LEVEL", "info")
	vipHook.SetDefault("HTTP_PORT", "8080")
	vipHook.SetDefault("POSTGRES_DSN", "postgresql://user:password@localhost:5432/releaseradar?sslmode=disable")
	vipHook.SetDefault("REDIS_ADDR", "localhost:6379")
	vipHook.SetDefault("GITHUB_TOKEN", "")
	vipHook.SetDefault("TELEGRAM_BOT_TOKEN", "")

	// Bind environment variables manually to avoid issues with hyphens if used in config names
	_ = vipHook.BindEnv("LOG_LEVEL")
	_ = vipHook.BindEnv("HTTP_PORT")
	_ = vipHook.BindEnv("POSTGRES_DSN")
	_ = vipHook.BindEnv("REDIS_ADDR")
	_ = vipHook.BindEnv("GITHUB_TOKEN")
	_ = vipHook.BindEnv("TELEGRAM_BOT_TOKEN")

	vipHook.ReadInConfig() // Read config file if exists (e.g., .env)
}

// @title ReleaseRadar API
// @version 1.0
// @description This is the API for ReleaseRadar, a GitHub releases tracker with Telegram notifications.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
func main() {
	// Initialize logger
	logger.InitLogger(viper.GetString("LOG_LEVEL"))
	log := logger.L()
	defer func() {
		_ = log.Sync() // Flushes buffer, if any
	}()

	// Initialize OpenTelemetry Tracer
	shutdownTracer := otel.InitTracer("api-service", "1.0.0")
	defer shutdownTracer()

	// Initialize database
	dbStore, err := persistence.NewPostgresStore(viper.GetString("POSTGRES_DSN"))
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: viper.GetString("REDIS_ADDR"),
	})

	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal("failed to connect to redis", zap.Error(err))
	}
	// _ = cache.NewRedisCache(redisClient) // Удалена неиспользуемая переменная redisCache
	redisIdempotencyStorage := persistence.NewRedisIdempotencyStorage(redisClient)
	idempotencyManager := idempotency.NewManager(redisIdempotencyStorage)

	// Initialize GitHub client
	githubClient := github.NewGitHubClient(http.DefaultClient) // TODO: Add proper HTTP client with rate limit handling

	// Initialize Telegram client
	telegramClient, err := telegram.NewTelegramClient(viper.GetString("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		log.Fatal("failed to create telegram client", zap.Error(err))
	}

	// Initialize usecases
	_ = usecase.NewUserUseCase(dbStore, dbStore)                                                           // Удалена неиспользуемая переменная userUseCase
	_ = usecase.NewRepoUseCase(dbStore, dbStore, githubClient, dbStore)                                    // Удалена неиспользуемая переменная repoUseCase
	_ = usecase.NewSubscriptionUseCase(dbStore, dbStore, dbStore, dbStore)                                 // Удалена неиспользуемая переменная subscriptionUseCase
	_ = usecase.NewPollerUseCase(dbStore, dbStore, dbStore, dbStore, dbStore, githubClient, dbStore)       // Удалена неиспользуемая переменная pollerUseCase
	_ = usecase.NewNotifierUseCase(dbStore, dbStore, dbStore, telegramClient, idempotencyManager, dbStore) // Удалена неиспользуемая переменная notifierUseCase

	// _ = &usecase.Usecases{ // Удалена неиспользуемая переменная appUsecases
	// 	User:         userUseCase,
	// 	Repo:         repoUseCase,
	// 	Subscription: subscriptionUseCase,
	// 	Poller:       pollerUseCase,
	// 	Notifier:     notifierUseCase,
	// }

	// Set up Gin router
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logger.GinLogger(log))
	r.Use(RequestIDMiddleware())

	// Health check endpoint
	r.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	v1 := r.Group("/api/v1")
	{
		v1.POST("/signup", func(c *gin.Context) { /* signup stub */ })
		v1.POST("/repos", func(c *gin.Context) { /* add repo stub */ })
		v1.GET("/repos", func(c *gin.Context) { /* list repos stub */ })
		v1.POST("/repos/:repoID/subscribe", func(c *gin.Context) { /* subscribe stub */ })
	}

	httpPort := viper.GetString("HTTP_PORT")
	srv := &http.Server{
		Addr:    ":" + httpPort,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("listen: %s", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server with a timeout of 5 seconds.
	p := make(chan os.Signal, 1)
	signal.Notify(p, syscall.SIGINT, syscall.SIGTERM)

	<-p
	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", zap.Error(err))
	}

	log.Info("Server exiting")
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		uuid, _ := uuid.NewRandom()
		c.Request.Header.Set("X-Request-ID", uuid.String())
		c.Next()
	}
}
