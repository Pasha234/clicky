//go:build integration

package testkit

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
)

const (
	rabbitMQUser = "clicky"
	rabbitMQPass = "test_password"
)

// TestRabbitMQ is an isolated RabbitMQ broker for collector integration tests.
// URL uses the same AMQP 1.0 client connection format as the collector.
type TestRabbitMQ struct {
	URL      string
	sequence atomic.Uint64
	close    func(context.Context) error
}

// NewRabbitMQ starts a broker for one test and registers its cleanup.
func NewRabbitMQ(t testing.TB) *TestRabbitMQ {
	t.Helper()

	broker, err := NewRabbitMQForSuite(context.Background())
	if err != nil {
		t.Fatalf("start test RabbitMQ: %v", err)
	}

	t.Cleanup(func() {
		if err := broker.Close(context.Background()); err != nil {
			t.Errorf("close test RabbitMQ: %v", err)
		}
	})

	return broker
}

// NewRabbitMQForSuite starts one broker that can be shared by a package's
// integration tests. Call Close after m.Run when using it from TestMain.
func NewRabbitMQForSuite(ctx context.Context) (*TestRabbitMQ, error) {
	container, err := rabbitmq.Run(
		ctx,
		"rabbitmq:4-alpine",
		rabbitmq.WithAdminUsername(rabbitMQUser),
		rabbitmq.WithAdminPassword(rabbitMQPass),
	)
	if err != nil {
		return nil, fmt.Errorf("start RabbitMQ container: %w", err)
	}

	url, err := container.AmqpURL(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("get RabbitMQ URL: %w", err)
	}
	url += "/"

	if err := waitForRabbitMQ(ctx, url); err != nil {
		_ = container.Terminate(ctx)
		return nil, err
	}

	return &TestRabbitMQ{
		URL: url,
		close: func(ctx context.Context) error {
			return container.Terminate(ctx)
		},
	}, nil
}

// QueueName returns a fresh queue name. The broker does not declare it: queue
// declaration is part of the publisher behaviour being tested.
func (b *TestRabbitMQ) QueueName() string {
	return fmt.Sprintf("click_events_test_%d", b.sequence.Add(1))
}

// Close stops the RabbitMQ container.
func (b *TestRabbitMQ) Close(ctx context.Context) error {
	if b.close == nil {
		return nil
	}

	return b.close(ctx)
}

func waitForRabbitMQ(ctx context.Context, url string) error {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	for {
		env := rmq.NewEnvironment(url, nil)
		_, err := env.NewConnection(ctx)
		_ = env.CloseConnections(context.Background())
		if err == nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for RabbitMQ AMQP connection: %w", err)
		case <-time.After(100 * time.Millisecond):
		}
	}
}
