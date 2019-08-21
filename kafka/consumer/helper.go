package consumer

import (
	"strconv"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func stringPointer(src string) *string {
	return &src
}

func topicPartitionsToString(list []kafka.TopicPartition) string {

	var (
		topic string
		m     = make(map[string]*strings.Builder)
	)

	for _, p := range list {
		topic = ""
		if p.Topic != nil {
			topic = *p.Topic
		}

		partitions, ok := m[topic]
		if !ok {
			partitions = &strings.Builder{}
			m[topic] = partitions
		}

		if partitions.Len() > 0 {
			partitions.WriteByte(',')
		}

		partitions.WriteString(strconv.Itoa(int(p.Partition)))
	}

	b := strings.Builder{}
	for topic, partitions := range m {
		if b.Len() > 0 {
			b.WriteByte(',')
		}

		b.WriteString(topic)
		b.WriteRune('[')
		b.WriteString(partitions.String())
		b.WriteRune(']')
	}

	return b.String()
}
