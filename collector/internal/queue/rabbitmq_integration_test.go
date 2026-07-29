package queue

import (
	"clicky-go-collector/internal/config"
	"clicky-go-collector/internal/event"
	"clicky-go-collector/internal/testkit"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
)

var testRabbitMQ *testkit.TestRabbitMQ

func TestMain(m *testing.M) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		os.Exit(m.Run())
	}

	var err error
	testRabbitMQ, err = testkit.NewRabbitMQForSuite(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "start test RabbitMQ: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	closeCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := testRabbitMQ.Close(closeCtx); err != nil {
		fmt.Fprintf(os.Stderr, "close test RabbitMQ: %v\n", err)
		code = 1
	}

	os.Exit(code)
}

func TestRabbitMQPublisherPublish(t *testing.T) {
	ctx := context.Background()
	queueName := testRabbitMQ.QueueName()

	publisher, err := NewRabbitMQPublisher(ctx, config.RabbitMQ{
		URL:   testRabbitMQ.URL,
		Queue: queueName,
	})
	if err != nil {
		t.Fatalf("create publisher: %v", err)
	}
	t.Cleanup(publisher.Close)

	// Use a separate AMQP connection to prove that RabbitMQ received the message.
	env := rmq.NewEnvironment(testRabbitMQ.URL, nil)
	consumerConnection, err := env.NewConnection(ctx)
	if err != nil {
		t.Fatalf("create consumer connection: %v", err)
	}
	t.Cleanup(func() { _ = env.CloseConnections(context.Background()) })

	consumer, err := consumerConnection.NewConsumer(ctx, queueName, nil)
	if err != nil {
		t.Fatalf("create consumer: %v", err)
	}
	t.Cleanup(func() { _ = consumer.Close(context.Background()) })

	want := &event.Event{
		Token:     "clk_event_token",
		Type:      "click",
		URL:       "https://example.test/products/42",
		Referrer:  "https://search.example.test/",
		IP:        net.ParseIP("203.0.113.10"),
		UserAgent: "integration-test",
		Meta:      map[string]any{"button": "buy"},
		Timestamp: time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC),
		SiteID:    "11111111-1111-1111-1111-111111111111",
	}

	if err := publisher.Publish(ctx, want); err != nil {
		t.Fatalf("Publish() error: %v", err)
	}

	receiveCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	delivery, err := consumer.Receive(receiveCtx)
	if err != nil {
		t.Fatalf("receive published message: %v", err)
	}

	if err := delivery.Accept(ctx); err != nil {
		t.Errorf("acknowledge delivery: %v", err)
	}

	message := delivery.Message()
	if message.Header == nil || !message.Header.Durable {
		t.Error("received message is not durable")
	}

	var got event.Event
	if err := json.Unmarshal(message.GetData(), &got); err != nil {
		t.Fatalf("decode received event: %v", err)
	}

	assertEventSame(t, want, &got)
}

func TestRabbitMQPublisherDeclaresDeadLetterTopology(t *testing.T) {
	ctx := context.Background()
	queueName := testRabbitMQ.QueueName()

	publisher, err := NewRabbitMQPublisher(ctx, config.RabbitMQ{
		URL:   testRabbitMQ.URL,
		Queue: queueName,
	})
	if err != nil {
		t.Fatalf("create publisher: %v", err)
	}
	t.Cleanup(publisher.Close)

	env := rmq.NewEnvironment(testRabbitMQ.URL, nil)
	consumerConnection, err := env.NewConnection(ctx)
	if err != nil {
		t.Fatalf("create consumer connection: %v", err)
	}
	t.Cleanup(func() { _ = env.CloseConnections(context.Background()) })

	consumer, err := consumerConnection.NewConsumer(ctx, queueName, nil)
	if err != nil {
		t.Fatalf("create consumer: %v", err)
	}
	t.Cleanup(func() { _ = consumer.Close(context.Background()) })

	deadLetterConsumer, err := consumerConnection.NewConsumer(
		ctx, queueName+"_dead_letter",
		nil,
	)
	if err != nil {
		t.Fatalf("create dead letter consumer: %v", err)
	}
	t.Cleanup(func() { _ = deadLetterConsumer.Close(context.Background()) })

	want := &event.Event{
		Token:     "clk_event_token",
		Type:      "click",
		URL:       "https://example.test/products/42",
		Referrer:  "https://search.example.test/",
		IP:        net.ParseIP("203.0.113.10"),
		UserAgent: "integration-test",
		Meta:      map[string]any{"button": "buy"},
		Timestamp: time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC),
		SiteID:    "11111111-1111-1111-1111-111111111111",
	}

	if err := publisher.Publish(ctx, want); err != nil {
		t.Fatalf("Publish() error: %v", err)
	}

	receiveCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	delivery, err := consumer.Receive(receiveCtx)
	if err != nil {
		t.Fatalf("receive published message: %v", err)
	}

	if err := delivery.Discard(ctx, nil); err != nil {
		t.Errorf("discard delivery: %v", err)
	}

	receiveCtx, cancel = context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	delivery, err = deadLetterConsumer.Receive(receiveCtx)
	if err != nil {
		t.Fatalf("receive dead letter message: %v", err)
	}

	message := delivery.Message()
	if message.Header == nil || !message.Header.Durable {
		t.Error("received message is not durable")
	}

	if err := delivery.Accept(ctx); err != nil {
		t.Errorf("accept dead letter delivery: %v", err)
	}

	var got event.Event
	if err := json.Unmarshal(message.GetData(), &got); err != nil {
		t.Fatalf("decode received event: %v", err)
	}

	assertEventSame(t, want, &got)
}

func assertEventSame(t *testing.T, expected *event.Event, actual *event.Event) {
	t.Helper()

	if actual.Token != expected.Token {
		t.Errorf("Token = %q, expected %q", actual.Token, expected.Token)
	}
	if actual.SiteID != expected.SiteID {
		t.Errorf("SiteID = %q, expected %q", actual.SiteID, expected.SiteID)
	}
	if actual.Type != expected.Type {
		t.Errorf("Type = %q, expected %q", actual.Type, expected.Type)
	}
	if actual.URL != expected.URL {
		t.Errorf("URL = %q, expected %q", actual.URL, expected.URL)
	}
	if actual.IP.String() != expected.IP.String() {
		t.Errorf("IP = %q, expected %q", actual.IP, expected.IP)
	}
	if !actual.Timestamp.Equal(expected.Timestamp) {
		t.Errorf("Timestamp = %s, expected %s", actual.Timestamp, expected.Timestamp)
	}
}
