package producer

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/dialogs/dialog-go-lib/kafka/consumer"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestSync(t *testing.T) {

	var Topic = "test-producer-sync-" + strconv.Itoa(int(time.Now().Unix()))

	createTopic(t, Topic, 1, 1)
	defer removeTopic(t, Topic)

	chMessages := make(chan *kafka.Message)

	consumerCfg := &consumer.Config{
		OnError: func(_ context.Context, _ *zap.Logger, err error) {
			require.NoError(t, err)
		},
		OnProcess: func(_ context.Context, _ *zap.Logger, msg *kafka.Message, c *kafka.Consumer) error {
			chMessages <- msg
			return nil
		},
		Topics: []string{Topic},
		ConfigMap: &kafka.ConfigMap{
			"group.id":           "test",
			"bootstrap.servers":  getKafkaServers(),
			"auto.offset.reset":  "earliest",
			"session.timeout.ms": 6000,
		},
	}

	c, err := consumer.New(consumerCfg, newLogger(t))
	require.NoError(t, err)
	defer c.Stop()
	go func() { require.NoError(t, c.Start()) }()

	p, err := NewSyncProducer(&kafka.ConfigMap{
		"bootstrap.servers": getKafkaServers(),
	})
	require.NoError(t, err)

	require.NoError(t, p.Produce(context.Background(), &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &Topic,
			Partition: kafka.PartitionAny,
		},
		Value: []byte(Topic),
	}))

	res := <-chMessages

	require.WithinDuration(t, time.Now(), res.Timestamp, time.Minute)
	require.Equal(t,
		&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &Topic,
				Partition: 0,
			},
			Value:         []byte(Topic),
			Timestamp:     res.Timestamp,
			TimestampType: 1,
		},
		res)
}

func newLogger(t *testing.T) *zap.Logger {
	t.Helper()

	zapCfg := zap.NewProductionConfig()
	zapCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapLogger, err := zapCfg.Build()
	require.NoError(t, err)

	return zapLogger
}

func removeTopic(t *testing.T, topic ...string) {
	t.Helper()

	c := newAdminClient(t)
	results, err := c.DeleteTopics(context.Background(), topic)
	require.NoError(t, err)

	for _, res := range results {
		require.Equal(t, kafka.ErrNoError, res.Error.Code(), res.Error.String())
	}
}

func createTopic(t *testing.T, topic string, numParts, replicationFactor int) {
	t.Helper()

	c := newAdminClient(t)

	results, err := c.CreateTopics(
		context.Background(),
		[]kafka.TopicSpecification{{
			Topic:             topic,
			NumPartitions:     numParts,
			ReplicationFactor: replicationFactor}},
		// Admin options
		kafka.SetAdminOperationTimeout(time.Second*5))
	require.NoError(t, err)

	for _, res := range results {
		require.Equal(t, kafka.ErrNoError, res.Error.Code(), res.Error.String())
	}
}

func newAdminClient(t *testing.T) *kafka.AdminClient {
	t.Helper()

	c, err := kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers": getKafkaServers(),
	})
	require.NoError(t, err)

	return c
}

func getKafkaServers() string {
	return "localhost:9092"
}
