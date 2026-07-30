package app

import (
	"clicky-go-collector/internal/config"
	"clicky-go-collector/internal/http"
	"clicky-go-collector/internal/queue"
	"clicky-go-collector/internal/queue/metrics"
	"clicky-go-collector/internal/ratelimit"
	"clicky-go-collector/internal/token"
	"context"
	"fmt"
	"log"
	"time"
)

type App struct {
}

func NewApp() *App {
	return &App{}
}

func (a *App) Run(ctx context.Context) error {
	cfg := config.Load()

	publisher, err := queue.NewRabbitMQPublisher(
		ctx,
		cfg.RabbitMQ,
	)
	if err != nil {
		return fmt.Errorf("create RabbitMQ publisher: %w", err)
	}
	defer publisher.Close()

	tokens, err := token.NewPostgresValidator(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("create token validator: %w", err)
	}
	defer tokens.Close()

	limiter, err := ratelimit.NewRedisLimiter(ctx, cfg.Redis, cfg.HTTP.RateLimitPerMinute)
	if err != nil {
		return fmt.Errorf("create Redis rate limiter: %w", err)
	}
	defer func() {
		if err := limiter.Close(); err != nil {
			log.Printf("close Redis limiter: %v", err)
		}
	}()

	metrics.Register()

	handler := http.NewHandler(publisher, tokens, http.Options{
		Limiter:        limiter,
		CORSOrigins:    cfg.HTTP.CORSOrigins,
		ProxyHeader:    cfg.HTTP.ProxyHeader,
		TrustedProxies: cfg.HTTP.TrustedProxies,
	})
	serverErrors := make(chan error, 1)
	go func() { serverErrors <- handler.Listen(":3000") }()

	select {
	case err := <-serverErrors:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := handler.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown HTTP server: %w", err)
		}

		return nil
	}
}
