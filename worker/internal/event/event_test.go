package event

import (
	"net"
	"testing"
	"time"
)

func TestEventValidate(t *testing.T) {
	valid := Event{
		Token:     "clk_test",
		SiteID:    "site-id",
		Type:      "click",
		URL:       "https://example.com",
		IP:        net.ParseIP("127.0.0.1"),
		Timestamp: time.Now(),
	}

	tests := []struct {
		name    string
		event   Event
		wantErr bool
	}{
		{
			name:  "valid event",
			event: valid,
		},
		{
			name: "missing token",
			event: func() Event {
				e := valid
				e.Token = ""
				return e
			}(),
			wantErr: true,
		},
		{
			name: "invalid URL",
			event: func() Event {
				e := valid
				e.URL = "not a URL"
				return e
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.event.Validate()

			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
