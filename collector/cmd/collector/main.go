package main

import (
	app2 "clicky-go-collector/internal/app"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	app := app2.NewApp()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
