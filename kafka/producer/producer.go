package producer

import (
	"context"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Producer interface {
	Produce(ctx context.Context, msg *kafka.Message) error
	Close()
}
