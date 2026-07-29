package http

import (
	"clicky-go-collector/internal/event"
	"clicky-go-collector/internal/token"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestOptionalUint16(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *uint16
		wantErr bool
	}{
		{
			name:  "empty value",
			input: "",
		},
		{
			name:  "valid number",
			input: "120",
			want: func() *uint16 {
				v := uint16(120)
				return &v
			}(),
		},
		{
			name:    "negative number",
			input:   "-1",
			wantErr: true,
		},
		{
			name:    "too large",
			input:   "70000",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := optionalUint16(tt.input)

			if (err != nil) != tt.wantErr {
				t.Fatalf("optionalUint16() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.want == nil && got != nil {
				t.Fatalf("optionalUint16() = %d, want nil", *got)
			}

			if tt.want != nil && (got == nil || *got != *tt.want) {
				t.Fatalf("optionalUint16() = %v, want %d", got, *tt.want)
			}
		})
	}
}

type fakePublisher struct {
	published    *event.Event
	err          error
	publishCalls int
}

func (p *fakePublisher) Publish(_ context.Context, received *event.Event) error {
	p.publishCalls++
	if p.err != nil {
		return p.err
	}
	copy := *received
	p.published = &copy

	return nil
}

func (p *fakePublisher) Ready(_ context.Context) error {
	return nil
}

func (v *fakePublisher) Clear() {
	v.published = nil
	v.publishCalls = 0
	v.err = nil
}

type fakeValidator struct {
	siteID string
	err    error
}

func (v *fakeValidator) Validate(_ context.Context, _ string) (string, error) {
	return v.siteID, v.err
}

func TestCollectPostRequest(t *testing.T) {
	publisher := &fakePublisher{}
	validator := &fakeValidator{
		siteID: "site",
	}

	tests := []struct {
		name             string
		headers          map[string]string
		input            event.Input
		inputJSON        string
		wantStatusCode   int
		wantPublished    bool
		validator        *fakeValidator
		publisher        *fakePublisher
		wantPublishCalls int
	}{
		{
			name: "valid event",
			headers: map[string]string{
				"User-Agent":   "collector-test",
				"Content-Type": "application/json",
			},
			input: event.Input{
				Token: "clk_test",
				Type:  "click",
				URL:   "https://example.com",
			},
			wantStatusCode: fiber.StatusAccepted,
			wantPublished:  true,
			validator:      validator,
			publisher:      publisher,
		},
		{
			name: "Invalid JSON",
			headers: map[string]string{
				"User-Agent":   "collector-test",
				"Content-Type": "application/json",
			},
			input: event.Input{
				Token: "clk_test",
				Type:  "click",
				URL:   "https://example.com",
			},
			inputJSON:      `{"token":`,
			wantStatusCode: fiber.StatusBadRequest,
			wantPublished:  false,
			validator:      validator,
			publisher:      publisher,
		},
		{
			name: "Missing event fields",
			headers: map[string]string{
				"User-Agent":   "collector-test",
				"Content-Type": "application/json",
			},
			input:          event.Input{},
			wantStatusCode: fiber.StatusBadRequest,
			wantPublished:  false,
			validator:      validator,
			publisher:      publisher,
		},
		{
			name: "Invalid Token",
			headers: map[string]string{
				"User-Agent":   "collector-test",
				"Content-Type": "application/json",
			},
			input: event.Input{
				Token: "clk_test",
				Type:  "click",
				URL:   "https://example.com",
			},
			wantStatusCode: fiber.StatusUnauthorized,
			wantPublished:  false,
			validator: func() *fakeValidator {
				validator := fakeValidator{
					err: token.ErrInvalid,
				}
				return &validator
			}(),
			publisher: publisher,
		},
		{
			name: "PostgreSQL unavailable",
			headers: map[string]string{
				"User-Agent":   "collector-test",
				"Content-Type": "application/json",
			},
			input: event.Input{
				Token: "clk_test",
				Type:  "click",
				URL:   "https://example.com",
			},
			wantStatusCode: fiber.StatusServiceUnavailable,
			wantPublished:  false,
			validator: func() *fakeValidator {
				validator := fakeValidator{
					err: errors.New("postgresql: connection refused"),
				}
				return &validator
			}(),
			publisher: publisher,
		},
		{
			name: "RabbitMQ unavailable",
			headers: map[string]string{
				"User-Agent":   "collector-test",
				"Content-Type": "application/json",
			},
			input: event.Input{
				Token: "clk_test",
				Type:  "click",
				URL:   "https://example.com",
			},
			wantStatusCode: fiber.StatusServiceUnavailable,
			wantPublished:  false,
			validator:      validator,
			publisher: func() *fakePublisher {
				publisher := fakePublisher{
					err: errors.New("rabbitmq: connection refused"),
				}
				return &publisher
			}(),
			wantPublishCalls: 1,
		},
		{
			name: "Invalid x/y coordinate",
			headers: map[string]string{
				"User-Agent":   "collector-test",
				"Content-Type": "application/json",
			},
			inputJSON: `{
				"token": "clk_test",
				"event": "click",
				"url": "https://example.com",
				"x": -1,
				"y": -1
			}`,
			wantStatusCode: fiber.StatusBadRequest,
			wantPublished:  false,
			validator:      validator,
			publisher:      publisher,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(tt.publisher, tt.validator)
			body := ""
			if tt.inputJSON != "" {
				body = tt.inputJSON
			} else {
				encoded, err := json.Marshal(tt.input)
				if err != nil {
					t.Fatalf("json.Marshal() error = %v", err)
				}
				body = string(encoded)
			}
			request := httptest.NewRequest(
				http.MethodPost,
				"/collect",
				strings.NewReader(string(body)),
			)
			for key, value := range tt.headers {
				request.Header.Set(key, value)
			}
			request.RemoteAddr = "127.0.0.1"
			response, err := handler.fiber.Test(request)
			if err != nil {
				t.Fatalf("app.Test() error = %v", err)
			}

			if response.StatusCode != tt.wantStatusCode {
				t.Fatalf("status = %d, want %d", response.StatusCode, tt.wantStatusCode)
			}

			if tt.wantPublished {
				if tt.publisher.published == nil {
					t.Fatal("expected an event to be published")
				}

				if tt.publisher.published.SiteID != "site" {
					t.Fatalf(
						"published SiteID = %q, want %q",
						tt.publisher.published.SiteID,
						"site",
					)
				}
			} else if tt.publisher.published != nil {
				t.Fatalf(
					"expected no published event, got %#v",
					tt.publisher.published,
				)
			}

			if tt.wantPublishCalls > 0 && tt.publisher.publishCalls != tt.wantPublishCalls {
				t.Fatalf("publishCalls() = %d, want %d", tt.publisher.publishCalls, tt.wantPublishCalls)
			}

			tt.publisher.Clear()
		})
	}
}

func TestCollectGetRequest(t *testing.T) {
	publisher := &fakePublisher{}
	validator := &fakeValidator{
		siteID: "site",
	}
	handler := NewHandler(publisher, validator)

	query := url.Values{}
	query.Set("t", "clk_test")
	query.Set("event", "click")
	query.Set("url", "https://example.com")
	query.Set("referrer", "https://google.com")
	query.Set("x", "1")
	query.Set("y", "1")
	request := httptest.NewRequest(
		http.MethodGet,
		"/collect?"+query.Encode(),
		nil,
	)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "collector-test")
	request.RemoteAddr = "127.0.0.1"
	response, err := handler.fiber.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}

	if response.StatusCode != fiber.StatusAccepted {
		t.Fatalf("status = %d, want %d", response.StatusCode, fiber.StatusAccepted)
	}

	if publisher.published == nil {
		t.Fatal("expected an event to be published")
	}

	if publisher.published.SiteID != "site" {
		t.Fatalf(
			"published SiteID = %q, want %q",
			publisher.published.SiteID,
			"site",
		)
	}

	if publisher.published.X == nil {
		t.Fatal("published X is nil, want 1")
	}

	if *publisher.published.X != 1 {
		t.Fatalf("published X = %v, want %v", *publisher.published.X, 1)
	}

	if publisher.published.Y == nil {
		t.Fatal("published Y is nil, want 1")
	}

	if *publisher.published.Y != 1 {
		t.Fatalf("published Y = %v, want %v", *publisher.published.Y, 1)
	}
}
