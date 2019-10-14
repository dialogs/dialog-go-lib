package producer

import (
	"context"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type MockProducer struct {
	pipe chan *kafka.Message
}

func NewMockProducer() *MockProducer {
	return &MockProducer{
		pipe: make(chan *kafka.Message, 2),
	}
}

func (p *MockProducer) Produce(ctx context.Context, msg *kafka.Message) error {
	p.pipe <- msg
	return nil
}

func (p *MockProducer) Pipe() <-chan *kafka.Message {
	return p.pipe
}

func (p *MockProducer) Close() {
	close(p.pipe)
}
