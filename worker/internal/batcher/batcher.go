package batcher

import (
	"clicky-go-worker/internal/event"
	"context"
)

type Committer interface {
	Commit(ctx context.Context) error
}

type PendingEvent struct {
	Event     event.Event
	Committer Committer
}

type Batcher struct {
	pendingEvents []PendingEvent
	size          int
}

func New(size int) *Batcher {
	if size <= 0 {
		panic("batcher size must be greater than zero")
	}

	return &Batcher{
		size: size,
	}
}

func (b *Batcher) Add(e PendingEvent) bool {
	b.pendingEvents = append(b.pendingEvents, e)
	return len(b.pendingEvents) >= b.size
}

func (b *Batcher) Take() []PendingEvent {
	if len(b.pendingEvents) == 0 {
		return nil
	}

	events := b.pendingEvents
	b.pendingEvents = nil

	return events
}

func (b *Batcher) Len() int {
	return len(b.pendingEvents)
}
