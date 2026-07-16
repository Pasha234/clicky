package app

import (
	batcher2 "clicky-go-worker/internal/batcher"
	"clicky-go-worker/internal/config"
	"clicky-go-worker/internal/metrics"
	"clicky-go-worker/internal/queue"
	store2 "clicky-go-worker/internal/store"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type App struct {
}

func New() App {
	return App{}
}

func (app App) Start() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()
	cfg, err := config.Load()
	metrics.Register()

	if err != nil {
		panic(err)
	}

	receiver, err := queue.NewRabbitMQReceiver(
		ctx,
		cfg.RabbitMQURL,
		cfg.Queue,
		cfg.MaxInsertAttempts,
		cfg.RetryInitialDelay,
		cfg.RetryMaxDelay,
	)

	if err != nil {
		panic(err)
	}

	defer func() {
		closeCtx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		if err := receiver.Close(closeCtx); err != nil {
			log.Printf("close RabbitMQ receiver: %v", err)
		}
	}()

	go monitorQueueLag(ctx, receiver, 10*time.Second)

	batcher := batcher2.New(cfg.BatchSize)

	store, err := store2.NewClickHouseEventStore(ctx, cfg.ClickhouseDSN)

	if err != nil {
		panic(err)
	}

	metricsServer := startMetricsServer()
	defer metricsServer.Close()

	err = receiver.Run(ctx, batcher, store, cfg.FlushInterval)

	if err != nil {
		panic(err)
	}
}

func startMetricsServer() *http.Server {
	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := &http.Server{
		Addr:    ":3001",
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			log.Printf("metrics server failed: %v", err)
		}
	}()

	return server
}

func monitorQueueLag(
	ctx context.Context,
	receiver *queue.RabbitMQReceiver,
	interval time.Duration,
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		checkCtx, cancel := context.WithTimeout(ctx, 3*time.Second)

		lag, err := receiver.QueueLag(checkCtx)
		cancel()

		if err != nil {
			log.Printf("read RabbitMQ queue lag: %v", err)
		} else {
			metrics.QueueLag.Set(float64(lag))
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
