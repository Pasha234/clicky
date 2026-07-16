package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	EventsConsumed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "worker_events_consumed_total",
			Help: "Total RabbitMQ messages received by the worker.",
		},
	)

	EventsInserted = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "worker_events_inserted_total",
			Help: "Total events successfully inserted into ClickHouse.",
		},
	)

	EventsFailed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "worker_events_failed_total",
			Help: "Total events in batches that failed after all retries.",
		},
	)

	BatchSize = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name: "worker_batch_size",
			Help: "Number of events in each flushed batch.",
		},
	)

	BatchInsertDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name: "worker_batch_insert_duration_seconds",
			Help: "Time spent inserting a batch into ClickHouse.",
		},
	)

	ClickHouseErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "worker_clickhouse_errors_total",
			Help: "Total failed ClickHouse insert attempts.",
		},
	)

	QueueLag = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "worker_rabbitmq_queue_lag",
			Help: "Messages currently waiting in the RabbitMQ events queue.",
		},
	)
)

func Register() {
	prometheus.MustRegister(
		EventsConsumed,
		EventsInserted,
		EventsFailed,
		BatchSize,
		BatchInsertDuration,
		ClickHouseErrors,
		QueueLag,
	)
}
