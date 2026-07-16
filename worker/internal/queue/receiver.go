package queue

import (
	"clicky-go-worker/internal/batcher"
	"clicky-go-worker/internal/store"
	"context"
	"time"
)

type Receiver interface {
	Run(
		ctx context.Context,
		batcher *batcher.Batcher,
		store store.EventStore,
		flushInterval time.Duration,
	) error
	Ready(ctx context.Context) error
}
