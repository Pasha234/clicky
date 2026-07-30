package enrichment

import (
	"clicky-go-worker/internal/event"
	"net"

	"github.com/mssola/user_agent"
	"github.com/oschwald/geoip2-golang"
)

// Enricher derives analytics dimensions from fields already present in an
// event. GeoIP is optional: without a local MaxMind database, country and city
// are left empty while user-agent dimensions are still populated.
type Enricher struct {
	geoIP *geoip2.Reader
}

func New(geoIPDatabasePath string) (*Enricher, error) {
	if geoIPDatabasePath == "" {
		return &Enricher{}, nil
	}

	geoIP, err := geoip2.Open(geoIPDatabasePath)
	if err != nil {
		return nil, err
	}

	return &Enricher{geoIP: geoIP}, nil
}

func (e *Enricher) Close() error {
	if e == nil || e.geoIP == nil {
		return nil
	}

	return e.geoIP.Close()
}

func (e *Enricher) Enrich(value *event.Event) {
	if value == nil {
		return
	}

	e.enrichUserAgent(value)
	e.enrichLocation(value)
}

func (e *Enricher) enrichUserAgent(value *event.Event) {
	if value.UserAgent == "" {
		return
	}

	ua := user_agent.New(value.UserAgent)
	browser, _ := ua.Browser()
	value.Browser = browser
	value.OS = ua.OS()

	switch {
	case ua.Bot():
		value.Device = "Bot"
	case ua.Mobile():
		value.Device = "Mobile"
	default:
		value.Device = "Desktop"
	}
}

func (e *Enricher) enrichLocation(value *event.Event) {
	if e.geoIP == nil || value.IP == nil {
		return
	}

	location, err := e.geoIP.City(normalizeIP(value.IP))
	if err != nil {
		return
	}

	value.Country = location.Country.Names["en"]
	value.City = location.City.Names["en"]
}

func normalizeIP(ip net.IP) net.IP {
	if ipv4 := ip.To4(); ipv4 != nil {
		return ipv4
	}

	return ip
}
