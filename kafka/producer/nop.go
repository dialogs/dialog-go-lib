package producer

import (
	"context"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/zap"
)

type NopProducer struct {
	logger *zap.Logger
}

func NewNopProducer(logger *zap.Logger) Producer {
	return &NopProducer{
		logger: logger.With(zap.String("type", "no operation")),
	}
}

func (p *NopProducer) Produce(ctx context.Context, msg *kafka.Message) error {
	p.logger.Debug("write messages")
	return nil
}

func (p *NopProducer) Close() {
	// nothing do
}
