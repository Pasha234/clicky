package enrichment

import (
	"clicky-go-worker/internal/event"
	"net"
	"testing"
)

func TestEnricherEnrichesUserAgentWithoutGeoIPDatabase(t *testing.T) {
	enricher, err := New("")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	value := event.Event{
		IP: net.ParseIP("203.0.113.10"),
		UserAgent: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 " +
			"(KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}

	enricher.Enrich(&value)

	if value.Browser == "" {
		t.Error("Browser is empty")
	}
	if value.OS == "" {
		t.Error("OS is empty")
	}
	if value.Device != "Desktop" {
		t.Errorf("Device = %q, want Desktop", value.Device)
	}
	if value.Country != "" || value.City != "" {
		t.Errorf("location = (%q, %q), want empty without a GeoIP database", value.Country, value.City)
	}
}
