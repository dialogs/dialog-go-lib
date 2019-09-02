package consumer

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func stringPointer(src string) *string {
	return &src
}

func checkPartitions(list []kafka.TopicPartition) error {

	for i := range list {
		item := &list[i]
		if item.Error != nil {
			return item.Error
		}
	}

	return nil
}
