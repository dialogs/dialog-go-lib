package consumer

import (
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type ConsumerI interface {
	Delay(time.Duration, []kafka.TopicPartition) error
}
