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
)

func Register() {
	prometheus.MustRegister(
		RequestsTotal,
		RequestDuration,
		InvalidEvents,
		QueuePublishFailures,
	)
}
