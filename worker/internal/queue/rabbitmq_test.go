package queue

import (
	"clicky-go-worker/internal/event"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Azure/go-amqp"
	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
)

type fakeStore struct {
	results []error
	calls   int
}

func (s *fakeStore) Insert(_ context.Context, _ []event.Event) error {
	s.calls++

	result := s.results[0]
	s.results = s.results[1:]

	return result
}

type fakeDelivery struct {
	rmq.IDeliveryContext

	acceptCalls  int
	discardCalls int
	jsonEvent    []byte
}

func (d *fakeDelivery) Message() *amqp.Message {
	return &amqp.Message{
		Data: [][]byte{
			d.jsonEvent,
		},
	}
}

func (d *fakeDelivery) Accept(_ context.Context) error {
	d.acceptCalls++
	return nil
}
func (d *fakeDelivery) Discard(_ context.Context, _ *amqp.Error) error {
	d.discardCalls++

	return nil
}

func TestInsertWithRetryRetriesThenSucceeds(t *testing.T) {
	store := &fakeStore{
		results: []error{
			errors.New("ClickHouse unavailable"),
			errors.New("ClickHouse unavailable"),
			nil,
		},
	}

	err := insertWithRetry(
		context.Background(),
		store,
		[]event.Event{{Token: "clk_test"}},
		3,
		time.Nanosecond,
		time.Nanosecond,
	)

	if err != nil {
		t.Fatalf("insertWithRetry() error = %v, want nil", err)
	}

	if store.calls != 3 {
		t.Fatalf("Insert() calls = %d, want 3", store.calls)
	}
}

func TestDecodeDeliveryFails(t *testing.T) {
	event := event.Event{
		Token: "wrong_event",
	}
	eventJSON, err := event.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v, want nil", err)
	}
	delivery := fakeDelivery{
		jsonEvent: eventJSON,
	}

	_, err = decodeDelivery(
		context.Background(),
		&delivery,
	)

	if err == nil {
		t.Fatalf("decodeDelivery() error = nil, want non-nil")
	}

	if delivery.discardCalls != 1 {
		t.Fatalf("delivery.discardCalls() calls = %d, want 1", delivery.discardCalls)
	}
}
