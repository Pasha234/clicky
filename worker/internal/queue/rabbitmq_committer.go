package queue

import (
	"context"

	"github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
)

type RabbitMQCommitter struct {
	delivery rabbitmqamqp.IDeliveryContext
}

func (r RabbitMQCommitter) Commit(ctx context.Context) error {
	return r.delivery.Accept(ctx)
}
