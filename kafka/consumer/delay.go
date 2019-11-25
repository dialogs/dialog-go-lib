package consumer

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type DelayI interface {
	DelayConsumer(sec time.Duration) error
}

type Delay struct {
	consumer *kafka.Consumer
	ctx      context.Context
	logger   *zap.Logger
}

func NewDelay(ctx context.Context, consumer *kafka.Consumer, l *zap.Logger) Delay {
	return Delay{
		consumer: consumer,
		ctx:      ctx,
		logger:   l,
	}
}

func (d Delay) DelayConsumer(sleep time.Duration) error {
	partitions, err := d.consumer.Assignment()
	if err != nil {
		return err
	}
	err = d.consumer.Pause(partitions)
	if err != nil {
		return err
	}
	time.Sleep(sleep)
	select {
	case <-d.ctx.Done():
		d.logger.Info("service already stopped")
		return nil
	default:
		partitions, err = d.consumer.Assignment()
		if err != nil {
			return err
		}
		err = d.consumer.Resume(partitions)
		if err != nil {
			return err
		}
	}

	return nil
}
