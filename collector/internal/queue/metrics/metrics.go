package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "collector_requests_total",
			Help: "Total HTTP requests handled by the collector.",
		},
		[]string{"method", "status"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "collector_request_duration_seconds",
			Help: "Time spent handling collector requests.",
		},
		[]string{"method"},
	)

	InvalidEvents = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "collector_invalid_events_total",
			Help: "Total rejected invalid events.",
		},
	)

	QueuePublishFailures = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "collector_queue_publish_failures_total",
			Help: "Total RabbitMQ publish failures.",
		},
	)

	QueuePublishDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name: "collector_queue_publish_duration_seconds",
			Help: "Time spent publishing an event to RabbitMQ.",
		},
	)

	RateLimitedRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "collector_rate_limited_requests_total",
			Help: "Total collector requests rejected by rate limiting.",
		},
		[]string{"dimension"},
	)
)

func Register() {
	prometheus.MustRegister(
		RequestsTotal,
		RequestDuration,
		InvalidEvents,
		QueuePublishFailures,
		QueuePublishDuration,
		RateLimitedRequests,
	)
}
