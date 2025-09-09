package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// "fmt"

	"github.com/mackb/releaseradar/internal/adapter/github"
	"github.com/mackb/releaseradar/internal/adapter/persistence"
	"github.com/mackb/releaseradar/internal/adapter/telegram"
	"github.com/mackb/releaseradar/internal/usecase"
	"github.com/mackb/releaseradar/pkg/idempotency"
	"github.com/mackb/releaseradar/pkg/logger"
	"github.com/mackb/releaseradar/pkg/otel"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func init() {
	vipHook := viper.New()
	vipHook.AutomaticEnv()
	vipHook.SetEnvPrefix("RR_WORKER")
	vipHook.SetDefault("LOG_LEVEL", "info")
	vipHook.SetDefault("POSTGRES_DSN", "postgresql://user:password@localhost:5432/releaseradar?sslmode=disable")
	vipHook.SetDefault("REDIS_ADDR", "localhost:6379")
	vipHook.SetDefault("GITHUB_TOKEN", "")
	vipHook.SetDefault("TELEGRAM_BOT_TOKEN", "")
	vipHook.SetDefault("POLLER_INTERVAL_MINUTES", 5)
	vipHook.SetDefault("NOTIFIER_INTERVAL_SECONDS", 10)

	_ = vipHook.BindEnv("LOG_LEVEL")
	_ = vipHook.BindEnv("POSTGRES_DSN")
	_ = vipHook.BindEnv("REDIS_ADDR")
	_ = vipHook.BindEnv("GITHUB_TOKEN")
	_ = vipHook.BindEnv("TELEGRAM_BOT_TOKEN")
	_ = vipHook.BindEnv("POLLER_INTERVAL_MINUTES")
	_ = vipHook.BindEnv("NOTIFIER_INTERVAL_SECONDS")

	vipHook.ReadInConfig()
}

func main() {
	logger.InitLogger(viper.GetString("LOG_LEVEL"))
	log := logger.L()
	defer func() {
		_ = log.Sync()
	}()

	shutdownTracer := otel.InitTracer("worker-service", "1.0.0")
	defer shutdownTracer()

	dbStore, err := persistence.NewPostgresStore(viper.GetString("POSTGRES_DSN"))
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}

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

	githubClient := github.NewGitHubClient(http.DefaultClient)

	telegramClient, err := telegram.NewTelegramClient(viper.GetString("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		log.Fatal("failed to create telegram client", zap.Error(err))
	}

	pollerUseCase := usecase.NewPollerUseCase(dbStore, dbStore, dbStore, dbStore, dbStore, githubClient, dbStore)         // Обновленный вызов
	notifierUseCase := usecase.NewNotifierUseCase(dbStore, dbStore, dbStore, telegramClient, idempotencyManager, dbStore) // Обновленный вызов

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Poller loop
	pollerInterval := time.Duration(viper.GetInt("POLLER_INTERVAL_MINUTES")) * time.Minute
	go func() {
		ticker := time.NewTicker(pollerInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				log.Info("Running poller use case")
				err := pollerUseCase.PollReleases(ctx)
				if err != nil {
					log.Error("Poller use case failed", zap.Error(err))
				}
			case <-ctx.Done():
				log.Info("Poller stopped")
				return
			}
		}
	}()

	// Notifier loop
	notifierInterval := time.Duration(viper.GetInt("NOTIFIER_INTERVAL_SECONDS")) * time.Second
	go func() {
		ticker := time.NewTicker(notifierInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				log.Info("Running notifier use case")
				err := notifierUseCase.Notify(ctx)
				if err != nil {
					log.Error("Notifier use case failed", zap.Error(err))
				}
			case <-ctx.Done():
				log.Info("Notifier stopped")
				return
			}
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	log.Info("Shutting down worker...")
	cancel()
	// Give some time for goroutines to finish
	time.Sleep(2 * time.Second)
	log.Info("Worker exited")
}
