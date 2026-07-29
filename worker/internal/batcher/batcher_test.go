package batcher

import (
	"clicky-go-worker/internal/event"
	"testing"
)

func TestBatcherAddAndTake(t *testing.T) {
	batcher := New(2)

	firstIsFull := batcher.Add(PendingEvent{
		Event: event.Event{Token: "first"},
	})

	if firstIsFull {
		t.Fatal("first event must not fill a batch of two")
	}

	secondIsFull := batcher.Add(PendingEvent{
		Event: event.Event{Token: "second"},
	})

	if !secondIsFull {
		t.Fatal("second event must fill a batch of two")
	}

	pending := batcher.Take()

	if len(pending) != 2 {
		t.Fatalf("pending length must be 2, but got %d", len(pending))
	}

	if batcher.Len() != 0 {
		t.Fatalf("batcher length must be 0, but got %d", batcher.Len())
	}
}
