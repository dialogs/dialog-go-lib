package consumer

import (
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/require"
)

func TestOffsetAdd(t *testing.T) {

	o := newOffset()

	for i, offsetVal := range []kafka.Offset{2, 1} {
		o.Add(kafka.TopicPartition{Topic: stringPointer("t1"), Partition: 1, Offset: offsetVal})
		require.Equal(t,
			map[string]map[int32]kafka.Offset{
				"t1": map[int32]kafka.Offset{1: 2},
			},
			o.topics)
		require.Equal(t, i+1, o.Counter())
	}

	counterBefore := o.Counter()
	for i, offsetVal := range []kafka.Offset{1, 2, 3, 4, 5} {
		o.Add(kafka.TopicPartition{Topic: stringPointer("t2"), Partition: 2, Offset: offsetVal})
		require.Equal(t,
			map[string]map[int32]kafka.Offset{
				"t1": map[int32]kafka.Offset{1: 2},
				"t2": map[int32]kafka.Offset{2: offsetVal},
			},
			o.topics)
		require.Equal(t, i+1+counterBefore, o.Counter())
	}
}

func TestOffsetSync(t *testing.T) {

	o := newOffset()
	o.Add(
		kafka.TopicPartition{Topic: stringPointer("t1"), Partition: 1, Offset: 1},
		kafka.TopicPartition{Topic: stringPointer("t1"), Partition: 2, Offset: 1},
		kafka.TopicPartition{Topic: stringPointer("t2"), Partition: 0, Offset: 1},
		kafka.TopicPartition{Topic: stringPointer("t2"), Partition: 1, Offset: 1},
		kafka.TopicPartition{Topic: stringPointer("t3"), Partition: 1, Offset: 1},
	)

	counterBefore := o.Counter()
	o.Sync(
		kafka.TopicPartition{Topic: stringPointer("t1"), Partition: 1, Offset: 1},
		kafka.TopicPartition{Topic: stringPointer("t2"), Partition: 0, Offset: 1},
		kafka.TopicPartition{Topic: stringPointer("t2"), Partition: 1, Offset: 1},
		kafka.TopicPartition{Topic: stringPointer("t5"), Partition: 2, Offset: 1},
	)

	require.Equal(t, counterBefore, o.Counter())
	require.Equal(t,
		map[string]map[int32]kafka.Offset{
			"t1": map[int32]kafka.Offset{1: 1},
			"t2": map[int32]kafka.Offset{0: 1, 1: 1},
		},
		o.topics)

	o.Sync()

	require.Equal(t, counterBefore, o.Counter())
	require.Equal(t,
		map[string]map[int32]kafka.Offset{},
		o.topics)
}

func TestOffsetGetRemove(t *testing.T) {

	src := []kafka.TopicPartition{
		kafka.TopicPartition{Topic: stringPointer("t1"), Partition: 0, Offset: 1},
		kafka.TopicPartition{Topic: stringPointer("t1"), Partition: 1, Offset: 2},
		kafka.TopicPartition{Topic: stringPointer("t2"), Partition: 0, Offset: 3},
		kafka.TopicPartition{Topic: stringPointer("t3"), Partition: 0, Offset: 4},
	}

	o := newOffset()
	o.Add(src...)

	counterBefore := o.Counter()
	require.Equal(t, len(src), counterBefore)

	for i, item := range src {

		partitions := o.Get()
		require.NotNil(t, partitions)
		require.Equal(t, counterBefore-i, o.Counter())

		o.Remove(item)
		require.Equal(t, counterBefore-(i+1), o.Counter())

		if i+1 < len(src) {
			require.NotEqual(t, map[string]map[int32]kafka.Offset{}, o.topics)
		}
	}

	require.Equal(t, 0, o.Counter())
	require.Equal(t, map[string]map[int32]kafka.Offset{}, o.topics)
}
