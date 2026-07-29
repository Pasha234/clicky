package event

import (
	"net"
	"testing"
	"time"
)

func TestNormalize(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		input     Input
		clientIP  string
		userAgent string
		wantErr   bool
	}{
		{
			name: "normalizes a valid event",
			input: Input{
				Token:    "token",
				Type:     "click",
				URL:      "http://localhost/event",
				Referrer: "http://localhost",
			},
			clientIP:  "127.0.0.1",
			userAgent: "Test browser",
		},
		{
			name: "rejects missing token",
			input: Input{
				Type: "click",
				URL:  "http://localhost/event",
			},
			clientIP:  "127.0.0.1",
			userAgent: "Test browser",
			wantErr:   true,
		},
		{
			name: "rejects invalid IP",
			input: Input{
				Token: "token",
				Type:  "click",
				URL:   "http://localhost/event",
			},
			clientIP:  "not-an-ip",
			userAgent: "Test browser",
			wantErr:   true,
		},
		{
			name: "rejects invalid timestamp",
			input: Input{
				Token:     "token",
				Type:      "click",
				URL:       "http://localhost/event",
				Timestamp: "not-an-timestamp",
			},
			clientIP: "127.0.0.1",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Normalize(tt.input, tt.clientIP, tt.userAgent, now)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Normalize() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if got.Token != "token" {
				t.Fatalf("Token = %q, want %q", got.Token, "token")
			}

			if got.Type != "click" {
				t.Fatalf("Type = %q, want %q", got.Type, "click")
			}

			if got.IP.String() != net.ParseIP(tt.clientIP).String() {
				t.Fatalf("IP = %q, want %q", got.IP, tt.clientIP)
			}

			if !got.Timestamp.Equal(now) {
				t.Fatalf("Timestamp = %s, want %s", got.Timestamp, now)
			}
		})
	}
}
