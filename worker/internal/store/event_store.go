package store

import (
	"clicky-go-worker/internal/event"
	"context"
)

type EventStore interface {
	Insert(ctx context.Context, events []event.Event) error
}
