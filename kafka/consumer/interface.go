package consumer

import (
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type ISleeper interface {
	Sleep(time.Duration, []kafka.TopicPartition) error
}
