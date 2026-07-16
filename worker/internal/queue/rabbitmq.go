package queue

import (
	"clicky-go-worker/internal/batcher"
	batcher2 "clicky-go-worker/internal/batcher"
	"clicky-go-worker/internal/event"
	event2 "clicky-go-worker/internal/event"
	"clicky-go-worker/internal/metrics"
	"clicky-go-worker/internal/store"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
)

type RabbitMQReceiver struct {
	conn         *rmq.AmqpConnection
	queue        string
	env          *rmq.Environment
	maxAttempts  int
	initialDelay time.Duration
	maxDelay     time.Duration
}

func NewRabbitMQReceiver(
	ctx context.Context,
	rabbitMQURL string,
	queue string,
	maxAttempts int,
	initialDelay time.Duration,
	maxDelay time.Duration,
) (*RabbitMQReceiver, error) {
	env := rmq.NewEnvironment(rabbitMQURL, nil)
	conn, err := env.NewConnection(ctx)
	if err != nil {
		return nil, fmt.Errorf("error creating rabbitmq connection: %w", err)
	}

	err = declareEventTopology(ctx, conn.Management(), queue)
	if err != nil {
		return nil, fmt.Errorf("error creating rabbitmq queue: %w", err)
	}

	return &RabbitMQReceiver{
		conn:         conn,
		env:          env,
		queue:        queue,
		maxAttempts:  maxAttempts,
		initialDelay: initialDelay,
		maxDelay:     maxDelay,
	}, nil
}

func (r *RabbitMQReceiver) Run(
	ctx context.Context,
	batcher *batcher.Batcher,
	store store.EventStore,
	flushInterval time.Duration,
) error {

	consumer, err := r.conn.NewConsumer(
		ctx,
		r.queue,
		nil,
	)
	if err != nil {
		return err
	}
	defer func() { _ = consumer.Close(context.Background()) }()

	type receiveResult struct {
		delivery rmq.IDeliveryContext
		err      error
	}

	received := make(chan receiveResult)

	go func() {
		defer close(received)

		for {
			delivery, err := consumer.Receive(ctx)

			select {
			case received <- receiveResult{
				delivery: delivery,
				err:      err,
			}:
			case <-ctx.Done():
				return
			}

			if err != nil {
				return
			}
		}
	}()

	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		case result, ok := <-received:
			if !ok {
				if ctx.Err() != nil {
					return flushOnShutdown(batcher, store, r)
				}
				return nil
			}

			if result.err != nil {
				if errors.Is(result.err, context.Canceled) {
					return flushOnShutdown(batcher, store, r)
				}

				return result.err
			}

			metrics.EventsConsumed.Inc()

			pendingEvent, err := decodeDelivery(ctx, result.delivery)
			if err != nil {
				continue
			}

			if batcher.Add(pendingEvent) {
				if err := flush(ctx, batcher, store, r); err != nil {
					return err
				}
			}

		case <-ticker.C:
			if err := flush(ctx, batcher, store, r); err != nil {
				return err
			}

		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(
				context.Background(),
				10*time.Second,
			)
			err := flush(shutdownCtx, batcher, store, r)
			cancel()

			return err
		}
	}
}

func (r *RabbitMQReceiver) Close(ctx context.Context) error {
	if r.conn != nil {
		if err := r.conn.Close(ctx); err != nil {
			return err
		}
	}
	if r.env != nil {
		if err := r.env.CloseConnections(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (r *RabbitMQReceiver) Ready(ctx context.Context) error {
	err := declareEventTopology(ctx, r.conn.Management(), r.queue)
	if err != nil {
		return fmt.Errorf("RabbitMQ is not ready %w", err)
	}

	return nil
}

func flush(
	ctx context.Context,
	b *batcher.Batcher,
	store store.EventStore,
	r *RabbitMQReceiver,
) error {
	pendingEvents := b.Take()

	if len(pendingEvents) == 0 {
		return nil
	}

	metrics.BatchSize.Observe(float64(len(pendingEvents)))

	events := make([]event.Event, len(pendingEvents))

	for i, pending := range pendingEvents {
		events[i] = pending.Event
	}

	started := time.Now()

	if err := insertWithRetry(
		ctx,
		store,
		events,
		r.maxAttempts,
		r.initialDelay,
		r.maxDelay,
	); err != nil {
		metrics.EventsFailed.Add(float64(len(events)))
		return err
	}

	metrics.BatchInsertDuration.Observe(time.Since(started).Seconds())
	metrics.EventsInserted.Add(float64(len(events)))

	for _, pending := range pendingEvents {
		if err := pending.Committer.Commit(ctx); err != nil {
			return err
		}
	}

	return nil
}

func flushOnShutdown(
	batcher *batcher.Batcher,
	store store.EventStore,
	r *RabbitMQReceiver,
) error {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	return flush(ctx, batcher, store, r)
}

func decodeDelivery(ctx context.Context, delivery rmq.IDeliveryContext) (batcher2.PendingEvent, error) {
	var event event2.Event
	msg := delivery.Message()
	if len(msg.Data) == 0 {
		if err := delivery.Discard(ctx, nil); err != nil {
			return batcher2.PendingEvent{}, err
		}
		return batcher2.PendingEvent{}, errors.New("delivery message has no data")
	}
	err := json.Unmarshal(msg.Data[0], &event)

	if err != nil {
		if delivery.Discard(ctx, nil); err != nil {
			return batcher2.PendingEvent{}, err
		}
		return batcher2.PendingEvent{}, fmt.Errorf("decode event %w", err)
	}
	if err := event.Validate(); err != nil {
		if discardErr := delivery.Discard(ctx, nil); discardErr != nil {
			return batcher2.PendingEvent{}, discardErr
		}

		return batcher2.PendingEvent{}, fmt.Errorf("validate event %w", err)
	}
	pendingEvent := batcher2.PendingEvent{
		Event: event,
		Committer: RabbitMQCommitter{
			delivery: delivery,
		},
	}

	return pendingEvent, nil
}

func insertWithRetry(
	ctx context.Context,
	store store.EventStore,
	events []event.Event,
	maxAttempts int,
	initialDelay time.Duration,
	maxDelay time.Duration,
) error {
	delay := initialDelay

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := store.Insert(ctx, events)
		if err == nil {
			return nil
		}

		metrics.ClickHouseErrors.Inc()

		if attempt == maxAttempts {
			return fmt.Errorf(
				"insert batch after %d attempts: %w",
				attempt,
				err,
			)
		}

		timer := time.NewTimer(delay)

		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()

		case <-timer.C:
		}

		delay *= 2

		if delay > maxDelay {
			delay = maxDelay
		}
	}

	return nil
}

func declareEventTopology(
	ctx context.Context,
	management *rmq.AmqpManagement,
	queueName string,
) error {
	deadLetterExchange := queueName + "_dlx"
	deadLetterQueue := queueName + "_dead_letter"
	deadLetterRoutingKey := queueName + ".dead"

	if _, err := management.DeclareExchange(
		ctx,
		&rmq.DirectExchangeSpecification{
			Name: deadLetterExchange,
		},
	); err != nil {
		return fmt.Errorf("declare dead-letter exchange: %w", err)
	}

	if _, err := management.DeclareQueue(
		ctx,
		&rmq.ClassicQueueSpecification{
			Name: deadLetterQueue,
		},
	); err != nil {
		return fmt.Errorf("declare dead-letter queue: %w", err)
	}

	if _, err := management.Bind(
		ctx,
		&rmq.ExchangeToQueueBindingSpecification{
			SourceExchange:   deadLetterExchange,
			DestinationQueue: deadLetterQueue,
			BindingKey:       deadLetterRoutingKey,
		},
	); err != nil {
		return fmt.Errorf("bind dead-letter queue: %w", err)
	}

	if _, err := management.DeclareQueue(
		ctx,
		&rmq.ClassicQueueSpecification{
			Name:                 queueName,
			DeadLetterExchange:   deadLetterExchange,
			DeadLetterRoutingKey: deadLetterRoutingKey,
		},
	); err != nil {
		return fmt.Errorf("declare events queue: %w", err)
	}

	return nil
}

func (r *RabbitMQReceiver) QueueLag(ctx context.Context) (uint64, error) {
	info, err := r.conn.Management().QueueInfo(ctx, r.queue)
	if err != nil {
		return 0, fmt.Errorf("get queue info: %w", err)
	}

	return info.MessageCount(), nil
}
