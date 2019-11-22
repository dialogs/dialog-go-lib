package consumer

import (
	"context"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGroup(t *testing.T) {

	var Topic = "test-revoke-" + strconv.Itoa(int(time.Now().Unix()))

	const (
		CountPartitions = 10
		CountMessages   = 100
	)

	createTopic(t, Topic, CountPartitions, 1)
	defer func() { removeTopic(t, Topic) }()

	onError := func(_ context.Context, _ *zap.Logger, err error) {
		require.NoError(t, err)
	}

	chMsg := make(chan *kafka.Message, CountMessages)
	onProcess := func(_ context.Context, _ *zap.Logger, msg *kafka.Message, c *kafka.Consumer) error {
		chMsg <- msg
		return nil
	}

	cfg := GroupConfig{
		Workers: CountPartitions,
		Config:  newConsumerConfig([]string{Topic}, nil, onError, onProcess, nil, nil, nil),
	}

	l := newLogger(t)
	group, err := NewGroup(cfg, l)
	require.NoError(t, err)
	defer group.Stop()

	go func() { require.NoError(t, group.Start()) }()

	go func(topic string) {
		p := newProducer(t, topic)
		defer p.Close()

		for i := 0; i < CountMessages; i++ {
			deliveryChan := make(chan kafka.Event)
			require.NoError(t, p.Produce(
				&kafka.Message{
					TopicPartition: kafka.TopicPartition{
						Topic:     &topic,
						Partition: kafka.PartitionAny,
					},
					Value: []byte(strconv.Itoa(i)),
				},
				deliveryChan))

			event := <-deliveryChan
			eventMessage, ok := event.(*kafka.Message)
			require.True(t, ok, "%#v", event)
			require.NoError(t, eventMessage.TopicPartition.Error, " %#v", event)
		}
	}(Topic)

	ids := make([]int, 0, CountMessages)
	for msg := range chMsg {
		msgID, err := strconv.Atoi(string(msg.Value))
		require.NoError(t, err)
		ids = append(ids, msgID)
		if len(ids) == cap(ids) {
			break
		}
	}

	select {
	case <-chMsg:
		t.Fatal("invalid value")
	default:
		// ok
	}

	sort.Slice(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})

	require.Len(t, ids, CountMessages)
	for i, item := range ids {
		require.Equal(t, i, item)
	}
}

func TestGroupGracefulShutdown(t *testing.T) {

	var Topic = "test-graceful-" + strconv.Itoa(int(time.Now().Unix()))

	const (
		CountPartitions = 100
		CountMessages   = 100
	)

	createTopic(t, Topic, CountPartitions, 1)
	defer func() { removeTopic(t, Topic) }()

	onError := func(_ context.Context, _ *zap.Logger, err error) {
		require.NoError(t, err)
	}

	chMsg := make(chan *kafka.Message, CountMessages)
	onProcess := func(_ context.Context, _ *zap.Logger, msg *kafka.Message, c *kafka.Consumer) error {
		chMsg <- msg
		return nil
	}

	cfg := GroupConfig{
		Workers: CountPartitions,
		Config:  newConsumerConfig([]string{Topic}, nil, onError, onProcess, nil, nil, nil),
	}

	l := newLogger(t)
	group, err := NewGroup(cfg, l)
	require.NoError(t, err)
	defer group.Stop()

	go func() { require.NoError(t, group.Start()) }()

	time.Sleep(time.Millisecond) // protection for error: 'consumers group already closed'
}
