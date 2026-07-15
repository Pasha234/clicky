package config

import "os"

const defaultBrokerURI = "amqp://clicky:clicky_local_password@127.0.0.1:5672/"
const defaultDatabaseURL = "postgres://clicky:clicky_local_password@127.0.0.1:6432/clicky?sslmode=disable&default_query_exec_mode=exec"

type Config struct {
	RabbitMQ RabbitMQ
	Database Database
}

type RabbitMQ struct {
	URL   string
	Queue string
}

type Database struct {
	URL string
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
	}
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
