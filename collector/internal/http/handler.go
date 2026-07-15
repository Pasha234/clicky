package http

import (
	"clicky-go-collector/internal/event"
	"clicky-go-collector/internal/queue"
	"clicky-go-collector/internal/queue/metrics"
	"clicky-go-collector/internal/token"
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	publisher queue.Publisher
	tokens    token.Validator
	fiber     *fiber.App
}

func NewHandler(publisher queue.Publisher, tokens token.Validator) *Handler {
	app := fiber.New()

	h := &Handler{
		publisher: publisher,
		tokens:    tokens,
		fiber:     app,
	}

	setRoutes(app, h)

	return h
}

func (h *Handler) Listen(addr string) error {
	return h.fiber.Listen(addr)
}

func (h *Handler) collectGet(c fiber.Ctx) error {
	x, err := optionalUint16(c.Query("x"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "x must be a number")
	}

	y, err := optionalUint16(c.Query("y"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "y must be a number")
	}

	input := event.Input{
		Token:     c.Query("t"),
		Type:      c.Query("event"),
		URL:       c.Query("url"),
		Referrer:  c.Query("referrer"),
		X:         x,
		Y:         y,
		Timestamp: c.Query("timestamp"),
	}

	return h.collect(c, input)
}

func (h *Handler) collectPost(c fiber.Ctx) error {
	var input event.Input

	if err := c.Bind().Body(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON")
	}

	return h.collect(c, input)
}

func (h *Handler) collect(c fiber.Ctx, input event.Input) error {
	started := time.Now()
	defer func() {
		metrics.RequestDuration.
			WithLabelValues(c.Method()).
			Observe(time.Since(started).Seconds())
	}()

	normalized, err := event.Normalize(
		input,
		c.IP(),
		c.Get("User-Agent"),
		time.Now(),
	)
	if err != nil {
		metrics.InvalidEvents.Inc()
		metrics.RequestsTotal.WithLabelValues(c.Method(), "400").Inc()
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	siteID, err := h.tokens.Validate(c.Context(), normalized.Token)

	switch {
	case errors.Is(err, token.ErrInvalid):
		return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
	case err != nil:
		return fiber.NewError(
			fiber.StatusServiceUnavailable,
			"token validation unavailable",
		)
	}

	normalized.SiteID = siteID

	if err := h.publisher.Publish(c.Context(), &normalized); err != nil {
		metrics.RequestsTotal.WithLabelValues(c.Method(), "503").Inc()
		return fiber.NewError(fiber.StatusServiceUnavailable, "queue unavailable")
	}

	metrics.RequestsTotal.WithLabelValues(c.Method(), "202").Inc()
	return c.SendStatus(fiber.StatusAccepted)
}

func optionalUint16(value string) (*uint16, error) {
	if value == "" {
		return nil, nil
	}

	n, err := strconv.ParseUint(value, 10, 16)
	if err != nil {
		return nil, err
	}

	result := uint16(n)
	return &result, nil
}
