//go:build integration

package http

import (
	"clicky-go-collector/internal/config"
	"clicky-go-collector/internal/queue"
	"clicky-go-collector/internal/testkit"
	"clicky-go-collector/internal/token"
	"context"
	"errors"
	"fmt"
	"net"
	stdhttp "net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
)

var (
	collectorTestDB       *testkit.TestDatabase
	collectorTestRabbitMQ *testkit.TestRabbitMQ
)

func TestMain(m *testing.M) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		os.Exit(m.Run())
	}

	var err error
	collectorTestDB, err = testkit.NewPostgresForSuite(
		context.Background(),
		"../token/testdata/schema.sql",
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start collector test database: %v\n", err)
		os.Exit(1)
	}

	collectorTestRabbitMQ, err = testkit.NewRabbitMQForSuite(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "start collector test RabbitMQ: %v\n", err)
		_ = collectorTestDB.Close(context.Background())
		os.Exit(1)
	}

	code := m.Run()
	closeCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := errors.Join(
		collectorTestRabbitMQ.Close(closeCtx),
		collectorTestDB.Close(closeCtx),
	); err != nil {
		fmt.Fprintf(os.Stderr, "close collector integration dependencies: %v\n", err)
		code = 1
	}

	os.Exit(code)
}

func TestCollectorRejectsInvalidTokenWithoutPublishing(t *testing.T) {
	requireCollectorIntegration(t)
	handler, queueName := newIntegrationHandler(t)

	request := httptest.NewRequest(stdhttp.MethodPost, "/collect", strings.NewReader(`{
		"token":"clk_unknown", "event":"click", "url":"https://example.test/"
	}`))
	request.Header.Set("Content-Type", "application/json")

	response, err := handler.fiber.Test(request)
	if err != nil {
		t.Fatalf("collector request: %v", err)
	}
	if response.StatusCode != stdhttp.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.StatusCode, stdhttp.StatusUnauthorized)
	}

	assertQueueHasNoMessage(t, queueName)
}

func TestCollectorRejectsOversizedRequestWithoutPublishing(t *testing.T) {
	requireCollectorIntegration(t)
	handler, queueName := newIntegrationHandler(t)
	body := `{"token":"clk_unknown","event":"click","url":"https://example.test/","meta":{"payload":"` +
		strings.Repeat("x", maxRequestBodySize) + `"}}`

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen for collector: %v", err)
	}
	serverDone := make(chan error, 1)
	go func() { serverDone <- handler.fiber.Listener(listener) }()
	t.Cleanup(func() {
		_ = handler.fiber.Shutdown()
		<-serverDone
	})

	request, err := stdhttp.NewRequest(
		stdhttp.MethodPost,
		"http://"+listener.Addr().String()+"/collect",
		strings.NewReader(body),
	)
	if err != nil {
		t.Fatalf("create oversized collector request: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := (&stdhttp.Client{Timeout: 5 * time.Second}).Do(request)
	if err != nil {
		t.Fatalf("collector request: %v", err)
	}
	t.Cleanup(func() { _ = response.Body.Close() })
	if response.StatusCode != stdhttp.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", response.StatusCode, stdhttp.StatusRequestEntityTooLarge)
	}

	assertQueueHasNoMessage(t, queueName)
}

func requireCollectorIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("set RUN_INTEGRATION_TESTS=1 to run integration tests")
	}
}

func newIntegrationHandler(t *testing.T) (*Handler, string) {
	t.Helper()
	ctx := context.Background()
	queueName := collectorTestRabbitMQ.QueueName()
	publisher, err := queue.NewRabbitMQPublisher(ctx, config.RabbitMQ{
		URL: collectorTestRabbitMQ.URL, Queue: queueName,
	})
	if err != nil {
		t.Fatalf("create RabbitMQ publisher: %v", err)
	}
	t.Cleanup(publisher.Close)

	validator, err := token.NewPostgresValidator(ctx, config.Database{URL: collectorTestDB.PgBouncerDSN})
	if err != nil {
		t.Fatalf("create token validator: %v", err)
	}
	t.Cleanup(validator.Close)

	return NewHandler(publisher, validator), queueName
}

func assertQueueHasNoMessage(t *testing.T, queueName string) {
	t.Helper()
	ctx := context.Background()
	env := rmq.NewEnvironment(collectorTestRabbitMQ.URL, nil)
	connection, err := env.NewConnection(ctx)
	if err != nil {
		t.Fatalf("create RabbitMQ inspection connection: %v", err)
	}
	t.Cleanup(func() { _ = env.CloseConnections(context.Background()) })

	consumer, err := connection.NewConsumer(ctx, queueName, nil)
	if err != nil {
		t.Fatalf("create RabbitMQ inspection consumer: %v", err)
	}
	t.Cleanup(func() { _ = consumer.Close(context.Background()) })

	receiveCtx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()
	if _, err := consumer.Receive(receiveCtx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected no RabbitMQ message, receive error = %v", err)
	}
}
