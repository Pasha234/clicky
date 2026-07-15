package queue

import (
	"clicky-go-collector/internal/event"
	"context"
)

type Publisher interface {
	Publish(ctx context.Context, event *event.Event) error
	Ready(ctx context.Context) error
}
