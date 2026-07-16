package main

import "clicky-go-worker/internal/app"

func main() {
	app := app.New()

	app.Start()
}
