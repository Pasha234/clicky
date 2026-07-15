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

type Input struct {
	Token     string         `json:"token"`
	Type      string         `json:"event"`
	URL       string         `json:"url"`
	Referrer  string         `json:"referrer"`
	X         *uint16        `json:"x"`
	Y         *uint16        `json:"y"`
	Meta      map[string]any `json:"meta"`
	Timestamp string         `json:"timestamp"`
}

func (e *Event) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// Normalize fills values that come from HTTP rather than the JSON/query input.
func Normalize(input Input, clientIP, userAgent string, now time.Time) (Event, error) {
	event := Event{
		Token:     strings.TrimSpace(input.Token),
		Type:      strings.ToLower(strings.TrimSpace(input.Type)),
		URL:       strings.TrimSpace(input.URL),
		Referrer:  strings.TrimSpace(input.Referrer),
		IP:        net.ParseIP(clientIP),
		UserAgent: userAgent,
		X:         input.X,
		Y:         input.Y,
		Meta:      input.Meta,
		Timestamp: now,
	}

	if input.Timestamp != "" {
		parsed, err := time.Parse(time.RFC3339, input.Timestamp)
		if err != nil {
			return Event{}, errors.New("timestamp must be RFC3339")
		}
		event.Timestamp = parsed
	}

	if event.Token == "" || event.Type == "" || event.URL == "" {
		return Event{}, errors.New("token, event, and url are required")
	}
	if _, err := url.ParseRequestURI(event.URL); err != nil {
		return Event{}, errors.New("url is invalid")
	}
	if event.IP == nil {
		return Event{}, errors.New("client IP is invalid")
	}

	return event, nil
}
