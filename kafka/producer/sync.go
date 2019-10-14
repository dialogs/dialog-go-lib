package producer

import (
	"context"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
)

type SyncProducer struct {
	producer *kafka.Producer
}

func NewSyncProducer(config *kafka.ConfigMap) (*SyncProducer, error) {
	producer, err := kafka.NewProducer(config)
	if err != nil {
		return nil, errors.Wrap(err, "create producer failed")
	}

	return &SyncProducer{
		producer: producer,
	}, nil
}

func (s *SyncProducer) Produce(ctx context.Context, msg *kafka.Message) error {
	delivery := make(chan kafka.Event, 1)

	err := s.producer.Produce(msg, delivery)
	if err != nil {
		return errors.Wrap(err, "produce message failed")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case e := <-delivery:
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				return errors.Wrap(ev.TopicPartition.Error, "produce message failed")
			}
			return nil
		case kafka.Error:
			return errors.Wrap(ev, "produce message failed")
		default:
			return errors.New("unknown produce event")
		}
	}
}

func (s *SyncProducer) Close() {
	s.producer.Close()
}
