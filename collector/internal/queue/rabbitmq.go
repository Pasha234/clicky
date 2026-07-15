package queue

import (
	"clicky-go-collector/internal/config"
	"clicky-go-collector/internal/event"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
)

type RabbitMQPublisher struct {
	connection *rmq.AmqpConnection
	publisher  *rmq.Publisher
	env        *rmq.Environment
	queue      string
}

func NewRabbitMQPublisher(
	ctx context.Context,
	cfg config.RabbitMQ,
) (*RabbitMQPublisher, error) {
	env := rmq.NewEnvironment(cfg.URL, nil)
	conn, err := env.NewConnection(ctx)
	if err != nil {
		return nil, fmt.Errorf("error creating rabbitmq connection: %w", err)
	}

	_, err = conn.Management().DeclareQueue(ctx, &rmq.QuorumQueueSpecification{Name: cfg.Queue})
	if err != nil {
		return nil, fmt.Errorf("error creating rabbitmq queue: %w", err)
	}

	publisher, err := conn.NewPublisher(ctx, &rmq.QueueAddress{Queue: cfg.Queue}, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating rabbitmq publisher: %w", err)
	}

	return &RabbitMQPublisher{
		publisher:  publisher,
		env:        env,
		queue:      cfg.Queue,
		connection: conn,
	}, nil
}

func (p *RabbitMQPublisher) Close() {
	if p.publisher != nil {
		_ = p.publisher.Close(context.Background())
	}
	if p.env != nil {
		_ = p.env.CloseConnections(context.Background())
	}
}

func (p *RabbitMQPublisher) Publish(ctx context.Context, event *event.Event) error {
	jsonStr, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		return err
	}
	res, err := p.publisher.Publish(ctx, rmq.NewMessage(jsonStr))
	if err != nil {
		log.Printf("Failed to publish message: %v", err)
		return err
	}

	switch res.Outcome.(type) {
	case *rmq.StateAccepted:
	default:
		log.Printf("Failed to publish message: %v", res.Outcome)
		return errors.New("Failed to publish message")
	}
	return nil
}

func (p *RabbitMQPublisher) Ready(ctx context.Context) error {
	_, err := p.connection.Management().DeclareQueue(
		ctx,
		&rmq.QuorumQueueSpecification{Name: p.queue},
	)
	if err != nil {
		return fmt.Errorf("RabbitMQ is not ready %w", err)
	}

	return nil
}
