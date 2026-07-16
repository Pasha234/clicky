// internal/event/event.go
package event

import (
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"strings"
	"time"
)

type Event struct {
	Token     string         `json:"token"`
	Type      string         `json:"event"`
	URL       string         `json:"url"`
	Referrer  string         `json:"referrer"`
	IP        net.IP         `json:"ip"`
	UserAgent string         `json:"user_agent"`
	X         *uint16        `json:"x"`
	Y         *uint16        `json:"y"`
	Meta      map[string]any `json:"meta"`
	Timestamp time.Time      `json:"timestamp"`
	SiteID    string         `json:"site_id"`
}

func (e *Event) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func (e Event) Validate() error {
	if strings.TrimSpace(e.Token) == "" {
		return errors.New("token is required")
	}

	if strings.TrimSpace(e.SiteID) == "" {
		return errors.New("site_id is required")
	}

	if strings.TrimSpace(e.Type) == "" {
		return errors.New("event is required")
	}

	if strings.TrimSpace(e.URL) == "" {
		return errors.New("url is required")
	}

	if _, err := url.ParseRequestURI(e.URL); err != nil {
		return errors.New("url is invalid")
	}

	if e.IP == nil {
		return errors.New("ip is required")
	}

	if e.Timestamp.IsZero() {
		return errors.New("timestamp is required")
	}

	return nil
}
