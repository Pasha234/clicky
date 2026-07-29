//go:build integration

package queue

import (
	"clicky-go-worker/internal/batcher"
	"clicky-go-worker/internal/event"
	"clicky-go-worker/internal/store"
	"clicky-go-worker/internal/testkit"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
)

var testRabbitMQ *testkit.TestRabbitMQ
var testClickHouse *testkit.TestClickHouse

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

	testClickHouse, err = testkit.NewClickHouseForSuite(
		context.Background(),
		"testdata/init-schema.sh",
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start test ClickHouse: %v\n", err)
		_ = testRabbitMQ.Close(context.Background())
		os.Exit(1)
	}

	code := m.Run()

	closeCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := testRabbitMQ.Close(closeCtx); err != nil {
		fmt.Fprintf(os.Stderr, "close test RabbitMQ: %v\n", err)
		code = 1
	}
	if err := testClickHouse.Close(closeCtx); err != nil {
		fmt.Fprintf(os.Stderr, "close test ClickHouse: %v\n", err)
		code = 1
	}

	os.Exit(code)
}

func TestRabbitMQReceiverRunInsertsAndAcknowledgesBatch(t *testing.T) {
	ctx := context.Background()
	queueName := testRabbitMQ.QueueName()
	clickhouseConn := initClickHouseConnection(t, ctx)
	defer clickhouseConn.Close()

	publisher := initRabbitMQPublisher(t, ctx, testRabbitMQ.URL, queueName)
	events := testkit.NewFakeEvents(2)

	receiver, err := NewRabbitMQReceiver(
		ctx,
		testRabbitMQ.URL,
		queueName,
		1,
		10*time.Millisecond,
		10*time.Millisecond,
	)
	if err != nil {
		t.Fatalf("create receiver: %v", err)
	}

	store, err := store.NewClickHouseEventStore(ctx, testClickHouse.DSN)
	if err != nil {
		t.Fatalf("create event store: %v", err)
	}

	runCtx, cancelRun := context.WithCancel(ctx)
	runDone := make(chan error, 1)
	go func() {
		runDone <- receiver.Run(runCtx, batcher.New(2), store, time.Hour)
	}()

	for i := range events {
		if err := publishEvent(publisher, ctx, &events[i]); err != nil {
			cancelRun()
			t.Fatalf("publish event %d: %v", i, err)
		}
	}

	waitForEventCount(t, ctx, clickhouseConn, len(events))

	// Run returns only after flush has inserted the batch and committed deliveries.
	cancelRun()
	select {
	case err := <-runDone:
		if err != nil {
			t.Fatalf("receiver Run() error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("receiver did not stop after context cancellation")
	}

	if err := receiver.Close(context.Background()); err != nil {
		t.Fatalf("close receiver: %v", err)
	}

	if lag := queueLag(t, ctx, testRabbitMQ.URL, queueName); lag != 0 {
		t.Errorf("queue lag after successful insert = %d, want 0 (events must be acknowledged)", lag)
	}
}

func TestRabbitMQReceiverRunFlushesOnInterval(t *testing.T) {
	ctx := context.Background()
	queueName := testRabbitMQ.QueueName()
	clickhouseConn := initClickHouseConnection(t, ctx)
	defer clickhouseConn.Close()

	publisher := initRabbitMQPublisher(t, ctx, testRabbitMQ.URL, queueName)
	want := testkit.NewFakeEvent()

	receiver, err := NewRabbitMQReceiver(
		ctx,
		testRabbitMQ.URL,
		queueName,
		1,
		10*time.Millisecond,
		10*time.Millisecond,
	)
	if err != nil {
		t.Fatalf("create receiver: %v", err)
	}

	store, err := store.NewClickHouseEventStore(ctx, testClickHouse.DSN)
	if err != nil {
		t.Fatalf("create event store: %v", err)
	}

	const flushInterval = 100 * time.Millisecond

	runCtx, cancelRun := context.WithCancel(ctx)
	runDone := make(chan error, 1)
	go func() {
		runDone <- receiver.Run(runCtx, batcher.New(2), store, flushInterval)
	}()

	if err := publishEvent(publisher, ctx, &want); err != nil {
		cancelRun()
		t.Fatalf("publish event: %v", err)
	}

	// BATCH_SIZE is 2 but only one event was published. Reaching ClickHouse
	// before cancellation proves that the ticker performed the flush.
	waitForEventCount(t, ctx, clickhouseConn, 1)

	cancelRun()
	select {
	case err := <-runDone:
		if err != nil {
			t.Fatalf("receiver Run() error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("receiver did not stop after context cancellation")
	}

	if err := receiver.Close(context.Background()); err != nil {
		t.Fatalf("close receiver: %v", err)
	}

	if lag := queueLag(t, ctx, testRabbitMQ.URL, queueName); lag != 0 {
		t.Errorf("queue lag after successful insert = %d, want 0 (events must be acknowledged)", lag)
	}
}

func TestRabbitMQReceiverRunDeadLettersInvalidEvent(t *testing.T) {
	ctx := context.Background()
	queueName := testRabbitMQ.QueueName()
	publisher := initRabbitMQPublisher(t, ctx, testRabbitMQ.URL, queueName)

	receiver, err := NewRabbitMQReceiver(
		ctx,
		testRabbitMQ.URL,
		queueName,
		1,
		10*time.Millisecond,
		10*time.Millisecond,
	)
	if err != nil {
		t.Fatalf("create receiver: %v", err)
	}

	eventStore, err := store.NewClickHouseEventStore(ctx, testClickHouse.DSN)
	if err != nil {
		t.Fatalf("create event store: %v", err)
	}

	runCtx, cancelRun := context.WithCancel(ctx)
	runDone := make(chan error, 1)
	go func() {
		runDone <- receiver.Run(runCtx, batcher.New(2), eventStore, time.Hour)
	}()

	deadLetterConsumer := initRabbitMQConsumer(t, ctx, testRabbitMQ.URL, queueName+"_dead_letter")
	invalidBody := []byte(`{"token":`)
	if err := publishMessage(publisher, ctx, invalidBody); err != nil {
		cancelRun()
		t.Fatalf("publish invalid event: %v", err)
	}

	delivery := receiveDelivery(t, ctx, deadLetterConsumer)
	if string(delivery.Message().GetData()) != string(invalidBody) {
		t.Errorf("dead-letter body = %q, want %q", delivery.Message().GetData(), invalidBody)
	}
	if err := delivery.Accept(ctx); err != nil {
		t.Errorf("acknowledge dead-letter delivery: %v", err)
	}

	stopReceiver(t, cancelRun, runDone, receiver)
}

func TestRabbitMQReceiverRunFlushesOnShutdown(t *testing.T) {
	ctx := context.Background()
	queueName := testRabbitMQ.QueueName()
	clickhouseConn := initClickHouseConnection(t, ctx)
	defer clickhouseConn.Close()

	publisher := initRabbitMQPublisher(t, ctx, testRabbitMQ.URL, queueName)
	want := testkit.NewFakeEvent()

	receiver, err := NewRabbitMQReceiver(
		ctx,
		testRabbitMQ.URL,
		queueName,
		1,
		10*time.Millisecond,
		10*time.Millisecond,
	)
	if err != nil {
		t.Fatalf("create receiver: %v", err)
	}

	eventStore, err := store.NewClickHouseEventStore(ctx, testClickHouse.DSN)
	if err != nil {
		t.Fatalf("create event store: %v", err)
	}

	pending := batcher.New(2)
	runCtx, cancelRun := context.WithCancel(ctx)
	runDone := make(chan error, 1)
	go func() {
		runDone <- receiver.Run(runCtx, pending, eventStore, time.Hour)
	}()

	if err := publishEvent(publisher, ctx, &want); err != nil {
		cancelRun()
		t.Fatalf("publish event: %v", err)
	}

	// The batch is incomplete and its timer is one hour, so it must still be
	// pending when shutdown starts.
	waitForBatchSize(t, pending, 1)
	assertEventCount(t, ctx, clickhouseConn, 0)

	stopReceiver(t, cancelRun, runDone, receiver)
	waitForEventCount(t, ctx, clickhouseConn, 1)

	if lag := queueLag(t, ctx, testRabbitMQ.URL, queueName); lag != 0 {
		t.Errorf("queue lag after shutdown flush = %d, want 0", lag)
	}
}

func TestClickHouseEventStoreInsert(t *testing.T) {
	ctx := context.Background()
	conn := initClickHouseConnection(t, ctx)
	defer conn.Close()

	eventStore, err := store.NewClickHouseEventStore(ctx, testClickHouse.DSN)
	if err != nil {
		t.Fatalf("create event store: %v", err)
	}

	want := testkit.NewFakeEvent()
	x, y := uint16(12), uint16(34)
	want.X = &x
	want.Y = &y

	if err := eventStore.Insert(ctx, []event.Event{want}); err != nil {
		t.Fatalf("Insert() error: %v", err)
	}

	var (
		siteID, token, eventType, url, referrer, userAgent, ip, meta string
		gotX, gotY                                                   uint16
		createdAt                                                    int64
	)
	err = conn.QueryRow(ctx, `
		SELECT
			toString(site_id), token, event_type, url, referrer,
			user_agent, toString(ip), ifNull(x, toUInt16(0)),
			ifNull(y, toUInt16(0)), meta, toUnixTimestamp64Milli(created_at)
		FROM events
	`).Scan(
		&siteID,
		&token,
		&eventType,
		&url,
		&referrer,
		&userAgent,
		&ip,
		&gotX,
		&gotY,
		&meta,
		&createdAt,
	)
	if err != nil {
		t.Fatalf("query inserted event: %v", err)
	}

	if siteID != want.SiteID {
		t.Errorf("site_id = %q, want %q", siteID, want.SiteID)
	}
	if token != want.Token {
		t.Errorf("token = %q, want %q", token, want.Token)
	}
	if eventType != want.Type {
		t.Errorf("event_type = %q, want %q", eventType, want.Type)
	}
	if url != want.URL || referrer != want.Referrer || userAgent != want.UserAgent {
		t.Errorf("stored request fields differ from event")
	}
	wantIP := want.IP.String()
	if want.IP.To4() != nil {
		wantIP = "::ffff:" + wantIP
	}
	if ip != wantIP {
		t.Errorf("ip = %q, want %q", ip, wantIP)
	}
	if gotX != x || gotY != y {
		t.Errorf("coordinates = (%d, %d), want (%d, %d)", gotX, gotY, x, y)
	}

	wantMeta, err := json.Marshal(want.Meta)
	if err != nil {
		t.Fatalf("marshal expected metadata: %v", err)
	}
	if meta != string(wantMeta) {
		t.Errorf("meta = %q, want %q", meta, wantMeta)
	}
	if createdAt != want.Timestamp.UnixMilli() {
		t.Errorf("created_at milliseconds = %d, want %d", createdAt, want.Timestamp.UnixMilli())
	}
}

func initClickHouseConnection(t *testing.T, ctx context.Context) driver.Conn {
	t.Helper()

	conn, err := testkit.ClickHouseConnect(ctx, testClickHouse.DSN)
	if err != nil {
		t.Fatalf("connect to ClickHouse: %v", err)
	}

	if err := conn.Exec(ctx, "TRUNCATE TABLE events"); err != nil {
		_ = conn.Close()
		t.Fatalf("clear ClickHouse events: %v", err)
	}

	return conn
}

func initRabbitMQPublisher(t *testing.T, ctx context.Context, url, queueName string) *rmq.Publisher {
	t.Helper()

	env := rmq.NewEnvironment(url, nil)
	conn, err := env.NewConnection(ctx)
	if err != nil {
		t.Fatalf("create publisher connection: %v", err)
	}
	t.Cleanup(func() { _ = env.CloseConnections(context.Background()) })

	if err := declareEventTopology(ctx, conn.Management(), queueName); err != nil {
		t.Fatalf("declare queue topology: %v", err)
	}

	publisher, err := conn.NewPublisher(ctx, &rmq.QueueAddress{Queue: queueName}, nil)
	if err != nil {
		t.Fatalf("create publisher: %v", err)
	}
	t.Cleanup(func() { _ = publisher.Close(context.Background()) })

	return publisher
}

func publishEvent(publisher *rmq.Publisher, ctx context.Context, event *event.Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return publishMessage(publisher, ctx, body)
}

func publishMessage(publisher *rmq.Publisher, ctx context.Context, body []byte) error {
	result, err := publisher.Publish(ctx, rmq.NewMessage(body))
	if err != nil {
		return err
	}

	if _, ok := result.Outcome.(*rmq.StateAccepted); !ok {
		return errors.New("RabbitMQ did not accept published message")
	}

	return nil
}

func initRabbitMQConsumer(t *testing.T, ctx context.Context, url, queueName string) *rmq.Consumer {
	t.Helper()

	env := rmq.NewEnvironment(url, nil)
	conn, err := env.NewConnection(ctx)
	if err != nil {
		t.Fatalf("create consumer connection: %v", err)
	}
	t.Cleanup(func() { _ = env.CloseConnections(context.Background()) })

	consumer, err := conn.NewConsumer(ctx, queueName, nil)
	if err != nil {
		t.Fatalf("create consumer: %v", err)
	}
	t.Cleanup(func() { _ = consumer.Close(context.Background()) })

	return consumer
}

func receiveDelivery(t *testing.T, ctx context.Context, consumer *rmq.Consumer) rmq.IDeliveryContext {
	t.Helper()

	receiveCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	delivery, err := consumer.Receive(receiveCtx)
	if err != nil {
		t.Fatalf("receive delivery: %v", err)
	}

	return delivery
}

func stopReceiver(
	t *testing.T,
	cancel context.CancelFunc,
	runDone <-chan error,
	receiver *RabbitMQReceiver,
) {
	t.Helper()
	cancel()

	select {
	case err := <-runDone:
		if err != nil {
			t.Fatalf("receiver Run() error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("receiver did not stop after context cancellation")
	}

	if err := receiver.Close(context.Background()); err != nil {
		t.Fatalf("close receiver: %v", err)
	}
}

func waitForEventCount(t *testing.T, ctx context.Context, conn driver.Conn, want int) {
	t.Helper()

	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		var got uint64
		err := conn.QueryRow(ctx, "SELECT count() FROM events").Scan(&got)
		if err == nil && got == uint64(want) {
			return
		}

		select {
		case <-deadline.C:
			if err != nil {
				t.Fatalf("query ClickHouse event count: %v", err)
			}
			t.Fatalf("ClickHouse event count = %d, want %d", got, want)
		case <-ticker.C:
		}
	}
}

func assertEventCount(t *testing.T, ctx context.Context, conn driver.Conn, want int) {
	t.Helper()

	var got uint64
	if err := conn.QueryRow(ctx, "SELECT count() FROM events").Scan(&got); err != nil {
		t.Fatalf("query ClickHouse event count: %v", err)
	}
	if got != uint64(want) {
		t.Fatalf("ClickHouse event count = %d, want %d", got, want)
	}
}

func waitForBatchSize(t *testing.T, b *batcher.Batcher, want int) {
	t.Helper()

	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		if b.Len() == want {
			return
		}

		select {
		case <-deadline.C:
			t.Fatalf("pending batch size = %d, want %d", b.Len(), want)
		case <-ticker.C:
		}
	}
}

func queueLag(t *testing.T, ctx context.Context, url, queueName string) uint64 {
	t.Helper()

	env := rmq.NewEnvironment(url, nil)
	conn, err := env.NewConnection(ctx)
	if err != nil {
		t.Fatalf("create queue-lag connection: %v", err)
	}
	t.Cleanup(func() { _ = env.CloseConnections(context.Background()) })

	info, err := conn.Management().QueueInfo(ctx, queueName)
	if err != nil {
		t.Fatalf("read queue lag: %v", err)
	}

	return info.MessageCount()
}
