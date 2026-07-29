//go:build integration

package testkit

import (
	"clicky-go-worker/internal/event"
	"fmt"
	"net"
	"time"
)

// NewFakeEvent returns one valid event suitable for worker tests.
func NewFakeEvent() event.Event {
	return newFakeEvent(1)
}

// NewFakeEvents returns distinct, valid events suitable for testing batches.
// Values are deterministic so a test failure is reproducible.
func NewFakeEvents(count int) []event.Event {
	events := make([]event.Event, count)

	for i := range events {
		events[i] = newFakeEvent(i + 1)
	}

	return events
}

func newFakeEvent(number int) event.Event {
	return event.Event{
		Token:     fmt.Sprintf("clk_test_token_%d", number),
		Type:      "click",
		URL:       fmt.Sprintf("https://example.test/products/%d", number),
		Referrer:  "https://search.example.test/",
		IP:        net.ParseIP("203.0.113.10"),
		UserAgent: "integration-test",
		Meta:      map[string]any{"button": "buy", "event_number": number},
		Timestamp: time.Date(2026, 7, 27, 12, 0, 0, number, time.UTC),
		SiteID:    "11111111-1111-1111-1111-111111111111",
	}
}
