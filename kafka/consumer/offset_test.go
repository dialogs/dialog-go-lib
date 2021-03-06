package consumer

import (
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/require"
)

func TestOffsetAdd(t *testing.T) {

	o := newOffset()

	for _, offsetVal := range []kafka.Offset{2, 1} {
		o.Add(kafka.TopicPartition{Topic: stringPointer("t1"), Partition: 1, Offset: offsetVal})
		require.Equal(t,
			map[string]map[int32]*offsetEntry{
				"t1": map[int32]*offsetEntry{1: &offsetEntry{Offset: 2, Count: 1}},
			},
			o.topics)
		require.Equal(t, 1, o.Counter())
	}

	counterBefore := o.Counter()
	for i, offsetVal := range []kafka.Offset{1, 2, 3, 4, 5} {
		o.Add(kafka.TopicPartition{Topic: stringPointer("t2"), Partition: 2, Offset: offsetVal})
		require.Equal(t,
			map[string]map[int32]*offsetEntry{
				"t1": map[int32]*offsetEntry{1: &offsetEntry{Offset: 2, Count: 1}},
				"t2": map[int32]*offsetEntry{2: &offsetEntry{Offset: offsetVal, Count: i + 1}},
			},
			o.topics)
		require.Equal(t, i+1+counterBefore, o.Counter())
	}
}

func TestOffsetClear(t *testing.T) {

	o := newOffset()
	o.Add(
		kafka.TopicPartition{Topic: stringPointer("t1"), Partition: 1, Offset: 1},
		kafka.TopicPartition{Topic: stringPointer("t1"), Partition: 2, Offset: 1},
		kafka.TopicPartition{Topic: stringPointer("t2"), Partition: 0, Offset: 1},
		kafka.TopicPartition{Topic: stringPointer("t2"), Partition: 1, Offset: 1},
		kafka.TopicPartition{Topic: stringPointer("t3"), Partition: 1, Offset: 1},
	)

	require.Equal(t, 5, o.Counter())

	o.Clear()

	require.Equal(t, 0, o.Counter())
	require.Equal(t,
		map[string]map[int32]*offsetEntry{},
		o.topics)
}

func TestOffsetGetRemove(t *testing.T) {

	src := []kafka.TopicPartition{
		kafka.TopicPartition{Topic: stringPointer("t1"), Partition: 0, Offset: 1},
		kafka.TopicPartition{Topic: stringPointer("t1"), Partition: 1, Offset: 2},
		kafka.TopicPartition{Topic: stringPointer("t1"), Partition: 1, Offset: 3},
		kafka.TopicPartition{Topic: stringPointer("t2"), Partition: 0, Offset: 3},
		kafka.TopicPartition{Topic: stringPointer("t3"), Partition: 0, Offset: 4},
	}

	o := newOffset()
	o.Add(src...)

	{
		_, count := o.Get()
		require.Equal(t, 1, count[getPartitionKey(stringPointer("t1"), 0)])
		require.Equal(t, 2, count[getPartitionKey(stringPointer("t1"), 1)])
		require.Equal(t, 1, count[getPartitionKey(stringPointer("t2"), 0)])
		require.Equal(t, 1, count[getPartitionKey(stringPointer("t3"), 0)])
	}

	counterBefore := o.Counter()
	require.Equal(t, len(src), counterBefore)

	for i, item := range src {

		partitions, count := o.Get()
		require.NotNil(t, partitions)
		var total int
		for _, v := range count {
			total += v
		}

		require.Equal(t, total, o.Counter())
		partitionCount := count[getPartitionKey(item.Topic, item.Partition)]

		o.Remove(item)
		counterBefore -= partitionCount
		require.Equal(t, counterBefore, o.Counter())

		if i+1 < len(src) {
			require.NotEqual(t, map[string]map[int32]kafka.Offset{}, o.topics)
		}
	}

	require.Equal(t, 0, o.Counter())
	require.Equal(t, map[string]map[int32]*offsetEntry{}, o.topics)
}
