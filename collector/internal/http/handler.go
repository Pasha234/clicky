package http

import (
	"clicky-go-collector/internal/event"
	"clicky-go-collector/internal/queue"
	"clicky-go-collector/internal/queue/metrics"
	"clicky-go-collector/internal/ratelimit"
	"clicky-go-collector/internal/token"
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	publisher queue.Publisher
	tokens    token.Validator
	limiter   ratelimit.Limiter
	fiber     *fiber.App
}

const maxRequestBodySize = 64 * 1024

type Options struct {
	Limiter        ratelimit.Limiter
	CORSOrigins    string
	ProxyHeader    string
	TrustedProxies []string
}

func NewHandler(publisher queue.Publisher, tokens token.Validator, options ...Options) *Handler {
	option := Options{
		Limiter:     ratelimit.AllowAll{},
		CORSOrigins: "*",
	}
	if len(options) > 0 {
		option = options[0]
		if option.Limiter == nil {
			option.Limiter = ratelimit.AllowAll{}
		}
		if option.CORSOrigins == "" {
			option.CORSOrigins = "*"
		}
	}
	app := fiber.New(fiber.Config{
		BodyLimit:          maxRequestBodySize,
		ProxyHeader:        option.ProxyHeader,
		TrustProxy:         len(option.TrustedProxies) > 0,
		EnableIPValidation: true,
		TrustProxyConfig: fiber.TrustProxyConfig{
			Proxies: option.TrustedProxies,
		},
	})

	h := &Handler{
		publisher: publisher,
		tokens:    tokens,
		limiter:   option.Limiter,
		fiber:     app,
	}

	setRoutes(app, h, option.CORSOrigins)

	return h
}

func (h *Handler) Listen(addr string) error {
	return h.fiber.Listen(addr)
}

func (h *Handler) Shutdown(ctx context.Context) error {
	return h.fiber.ShutdownWithContext(ctx)
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

	decision, err := h.limiter.Allow(c.Context(), normalized.Token, normalized.IP.String())
	if err != nil {
		metrics.RequestsTotal.WithLabelValues(c.Method(), "503").Inc()
		return fiber.NewError(fiber.StatusServiceUnavailable, "rate limiting unavailable")
	}
	if !decision.Allowed {
		metrics.RateLimitedRequests.WithLabelValues(decision.ExceededDimension).Inc()
		metrics.RequestsTotal.WithLabelValues(c.Method(), "429").Inc()
		return fiber.NewError(fiber.StatusTooManyRequests, "rate limit exceeded")
	}

	siteID, err := h.tokens.Validate(c.Context(), normalized.Token)

	switch {
	case errors.Is(err, token.ErrInvalid):
		metrics.RequestsTotal.WithLabelValues(c.Method(), "401").Inc()
		return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
	case err != nil:
		metrics.RequestsTotal.WithLabelValues(c.Method(), "503").Inc()
		return fiber.NewError(
			fiber.StatusServiceUnavailable,
			"token validation unavailable",
		)
	}

	normalized.SiteID = siteID

	publishStarted := time.Now()
	err = h.publisher.Publish(c.Context(), &normalized)
	metrics.QueuePublishDuration.Observe(time.Since(publishStarted).Seconds())
	if err != nil {
		metrics.QueuePublishFailures.Inc()
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
