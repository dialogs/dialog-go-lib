package consumer

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"time"
)

type DelayI interface {
	DelayConsumer(sec time.Duration) error
}

type Delay struct {
	consumer *kafka.Consumer
}

func NewDelay(consumer *kafka.Consumer) Delay{
	return Delay{consumer:consumer}
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
	err = d.consumer.Resume(partitions)
	if err != nil {
		return err
	}
	return nil
}

