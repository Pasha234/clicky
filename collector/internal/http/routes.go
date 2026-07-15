package http

import (
	"clicky-go-collector/internal/token"
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func setRoutes(app *fiber.App, h *Handler) {
	app.Get("/collect", h.collectGet)

	app.Post("/collect", h.collectPost)

	app.Get("/healthz", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	app.Get("/readyz", func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
		defer cancel()

		if err := h.publisher.Ready(ctx); err != nil {
			return fiber.NewError(
				fiber.StatusServiceUnavailable,
				"RabbitMQ is unavailable",
			)
		}

		if ready, ok := h.tokens.(token.Readiness); ok {
			if err := ready.Ready(ctx); err != nil {
				return fiber.NewError(
					fiber.StatusServiceUnavailable,
					"PostgreSQL is unavailable",
				)
			}
		}

		return c.SendStatus(fiber.StatusOK)
	})

	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
}
