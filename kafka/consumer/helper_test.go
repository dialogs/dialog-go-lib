package consumer

import (
	"errors"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/require"
)

func TestCheckPartitions(t *testing.T) {

	require.NoError(t, checkPartitions([]kafka.TopicPartition{}))

	require.NoError(t, checkPartitions([]kafka.TopicPartition{{}}))

	require.Equal(t,
		errors.New("fail"),
		checkPartitions([]kafka.TopicPartition{{Error: errors.New("fail")}}))
}
