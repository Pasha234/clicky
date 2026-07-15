package app

import (
	"clicky-go-collector/internal/config"
	"clicky-go-collector/internal/http"
	"clicky-go-collector/internal/queue"
	"clicky-go-collector/internal/queue/metrics"
	"clicky-go-collector/internal/token"
	"context"
	"log"
)

type App struct {
}

func NewApp() *App {
	return &App{}
}

func (a *App) Start() {
	cfg := config.Load()

	publisher, err := queue.NewRabbitMQPublisher(
		context.Background(),
		cfg.RabbitMQ,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer publisher.Close()

	tokens, err := token.NewPostgresValidator(context.Background(), cfg.Database)
	if err != nil {
		log.Fatal(err)
	}
	defer tokens.Close()

	metrics.Register()

	handler := http.NewHandler(publisher, tokens)

	error := handler.Listen(":3000")
	if error != nil {
		log.Fatal(error)
	}
}
