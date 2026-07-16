package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const defaultClickhouseDSN = "clickhouse://clicky:clicky_local_password@127.0.0.1:9000/clicky"
const defaultBrokerURI = "amqp://clicky:clicky_local_password@127.0.0.1:5672/"

type Config struct {
	RabbitMQURL       string
	Queue             string
	ClickhouseDSN     string
	BatchSize         int
	FlushInterval     time.Duration
	MaxInsertAttempts int
	RetryInitialDelay time.Duration
	RetryMaxDelay     time.Duration
}

func Load() (*Config, error) {
	queue := os.Getenv("RABBITMQ_QUEUE")
	if queue == "" {
		queue = "click_events"
	}

	batchSize, err := intEnv("BATCH_SIZE", 10)
	if err != nil {
		return &Config{}, err
	}

	flushInterval, err := durationEnv("FLUSH_INTERVAL", 10*time.Second)
	if err != nil {
		return &Config{}, err
	}

	maxInsertAttempts, err := intEnv("MAX_INSERT_ATTEMPTS", 10)
	if err != nil {
		return &Config{}, err
	}

	retryInitialDelay, err := durationEnv("RETRY_INITIAL_DELAY", 10*time.Second)
	if err != nil {
		return &Config{}, err
	}

	retryMaxDelay, err := durationEnv("RETRY_MAX_DELAY", 10*time.Second)
	if err != nil {
		return &Config{}, err
	}

	return &Config{
		RabbitMQURL:       brokerURI(),
		Queue:             queue,
		ClickhouseDSN:     clickhouseDSN(),
		BatchSize:         batchSize,
		FlushInterval:     flushInterval,
		MaxInsertAttempts: maxInsertAttempts,
		RetryInitialDelay: retryInitialDelay,
		RetryMaxDelay:     retryMaxDelay,
	}, nil
}

func brokerURI() string {
	uri := os.Getenv("RABBITMQ_URL")
	if uri == "" {
		uri = defaultBrokerURI
	}

	return uri
}

func clickhouseDSN() string {
	dsn := os.Getenv("CLICKHOUSE_DSN")
	if dsn == "" {
		dsn = defaultClickhouseDSN
	}

	return dsn
}

func intEnv(key string, defaultValue int) (int, error) {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue, nil
	}

	intVal, err := strconv.Atoi(val)

	if err != nil {
		return 0, err
	}

	if intVal <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}

	return intVal, nil
}

func durationEnv(key string, defaultValue time.Duration) (time.Duration, error) {
	value := os.Getenv(key)

	if value == "" {
		return defaultValue, nil
	}

	interval, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: %w", key, value, err)
	}

	if interval <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}

	return interval, nil
}
