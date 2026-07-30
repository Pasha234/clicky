package config

import (
	"os"
	"strconv"
	"strings"
)

const defaultBrokerURI = "amqp://clicky:clicky_local_password@127.0.0.1:5672/"
const defaultDatabaseURL = "postgres://clicky:clicky_local_password@127.0.0.1:6432/clicky?sslmode=disable&default_query_exec_mode=exec"
const defaultRedisURL = "redis://127.0.0.1:6379/0"

type Config struct {
	RabbitMQ RabbitMQ
	Database Database
	Redis    Redis
	HTTP     HTTP
}

type RabbitMQ struct {
	URL   string
	Queue string
}

type Database struct {
	URL string
}

type Redis struct {
	URL string
}

type HTTP struct {
	CORSOrigins        string
	RateLimitPerMinute int
	ProxyHeader        string
	TrustedProxies     []string
}

func Load() *Config {
	queue := os.Getenv("RABBITMQ_QUEUE")
	if queue == "" {
		queue = "click_events"
	}

	return &Config{
		RabbitMQ: RabbitMQ{
			URL:   brokerURI(),
			Queue: queue,
		},
		Database: Database{
			URL: databaseURL(),
		},
		Redis: Redis{
			URL: redisURL(),
		},
		HTTP: HTTP{
			CORSOrigins:        envOrDefault("COLLECTOR_CORS_ORIGINS", "*"),
			RateLimitPerMinute: positiveIntEnv("RATE_LIMIT_PER_MINUTE", 120),
			ProxyHeader:        envOrDefault("COLLECTOR_PROXY_HEADER", "X-Forwarded-For"),
			TrustedProxies:     commaSeparatedEnv("COLLECTOR_TRUSTED_PROXIES"),
		},
	}
}

func commaSeparatedEnv(name string) []string {
	value := os.Getenv(name)
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	proxies := make([]string, 0, len(parts))
	for _, part := range parts {
		if proxy := strings.TrimSpace(part); proxy != "" {
			proxies = append(proxies, proxy)
		}
	}

	return proxies
}

func redisURL() string {
	return envOrDefault("REDIS_URL", defaultRedisURL)
}

func envOrDefault(name, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}

	return value
}

func positiveIntEnv(name string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(name))
	if err != nil || value <= 0 {
		return fallback
	}

	return value
}

func brokerURI() string {
	uri := os.Getenv("RABBITMQ_URL")
	if uri == "" {
		uri = defaultBrokerURI
	}

	return uri
}

func databaseURL() string {
	uri := os.Getenv("DATABASE_URL")
	if uri == "" {
		uri = defaultDatabaseURL
	}

	return uri
}
