package consumer

import (
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/require"
)

func TestTopicPartitionsToString(t *testing.T) {

	require.Equal(t, "", topicPartitionsToString([]kafka.TopicPartition{}))

	require.Equal(t,
		"[1]",
		topicPartitionsToString([]kafka.TopicPartition{
			{Topic: nil, Partition: 1, Offset: 2, Error: nil},
		}))

	require.Equal(t,
		"topic-1[1]",
		topicPartitionsToString([]kafka.TopicPartition{
			{Topic: stringPointer("topic-1"), Partition: 1, Offset: 2, Error: nil},
		}))

	require.Contains(t,
		[]string{"topic-1[1,2],topic-2[1]", "topic-2[1],topic-1[1,2]"},
		topicPartitionsToString([]kafka.TopicPartition{
			{Topic: stringPointer("topic-1"), Partition: 1, Offset: 5, Error: nil},
			{Topic: stringPointer("topic-1"), Partition: 2, Offset: 5, Error: nil},
			{Topic: stringPointer("topic-2"), Partition: 1, Offset: 5, Error: nil},
		}))

}
