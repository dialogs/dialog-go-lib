package consumer

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"testing"
	"time"

	"go.uber.org/zap/zapcore"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func init() {
	log.SetFlags(log.Ltime | log.Lmicroseconds | log.Llongfile)
}

func TestConsumerNew(t *testing.T) {

	cfg := &Config{
		CommitOffsetCount:    1,
		CommitOffsetDuration: time.Hour,
		OnError:              func(context.Context, error) {},
		OnProcess:            func(context.Context, *kafka.Message) error { return nil },
		Topics:               []string{"a"},
		ConfigMap: &kafka.ConfigMap{
			"group.id":          "group-id",
			"bootstrap.servers": "b1,b2,b3",
		},
	}

	ctxDone, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := New(cfg, zap.L())
	require.NoError(t, err)
	require.Equal(t, ctxDone, c.ctx)
	require.Equal(t, zap.L(), c.logger)
	require.NotEmpty(t, c.onCommit, c.onCommit)
	require.NotEmpty(t, func(context.Context, error) {}, c.onError)
	require.NotEmpty(t, func(context.Context, *kafka.Message) {}, c.onProcess)
	require.Equal(t, []string{"a"}, c.topics)
	require.Equal(t, 1, c.commitOffsetCount)
	require.Equal(t, time.Hour, c.commitOffsetDuration)
	require.NotNil(t, c.reader)
}

func TestConsumerDoubleStartClose(t *testing.T) {

	var Topic = "test-doubleclose-" + strconv.Itoa(int(time.Now().Unix()))

	createTopic(t, Topic, 1, 1)
	defer removeTopic(t, Topic)

	onError := func(_ context.Context, err error) {
		require.NoError(t, err)
	}

	onProcess := func(_ context.Context, msg *kafka.Message) error {
		if msg == nil {
			return errors.New("invalid message")
		}
		return nil
	}

	onCommit := func(_ context.Context, topic string, partition int32, offset kafka.Offset) {
		require.Equal(t, Topic, topic)
		require.True(t, partition >= 0)
		require.True(t, offset >= 0)
	}

	c := newConsumer(t, Topic, nil, onError, onProcess, onCommit)
	go func() { require.NoError(t, c.Start()) }()
	defer func() { require.NoError(t, c.Stop()) }()

	time.Sleep(time.Second)
	require.NoError(t, c.Stop())

	require.EqualError(t, c.Start(), "consumer already closed")
}

func TestConsumerReadMessageSuccess(t *testing.T) {

	var Topic = "test-read-message-" + strconv.Itoa(int(time.Now().Unix()))

	createTopic(t, Topic, 1, 1)
	defer func() { removeTopic(t, Topic) }()

	onError := func(_ context.Context, err error) {
		require.NoError(t, err)
	}

	chMsg := make(chan *kafka.Message, 2)
	onProcess := func(_ context.Context, msg *kafka.Message) error {
		if msg == nil {
			return errors.New("invalid message")
		}
		chMsg <- msg
		return nil
	}

	onCommit := func(_ context.Context, topic string, partition int32, offset kafka.Offset) {
		require.Equal(t, Topic, topic)
		require.True(t, partition >= 0)
		require.True(t, offset >= 0)
	}

	c1 := newConsumer(t, Topic, nil, onError, onProcess, onCommit)
	defer func() { require.NoError(t, c1.Stop()) }()

	go func() { require.NoError(t, c1.Start()) }()

	p := newProducer(t, Topic)
	defer p.Close()

	deliveryChan := make(chan kafka.Event)
	require.NoError(t, p.Produce(
		&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &Topic,
				Partition: kafka.PartitionAny,
			},
			Value: []byte(Topic),
		},
		deliveryChan))
	require.Equal(t, 1, p.Flush(1))

	event := <-deliveryChan
	eventMessage, ok := event.(*kafka.Message)
	require.True(t, ok, "%#v", event)
	require.NoError(t, eventMessage.TopicPartition.Error, "%#v", event)

	c1.logger.Info("wait message")
	res := <-chMsg
	c1.logger.Info("read message")

	require.Equal(t, &Topic, res.TopicPartition.Topic)
	require.Equal(t, int32(0), res.TopicPartition.Partition)
	require.Equal(t, []byte(Topic), res.Value)

	require.NoError(t, c1.Stop())
}

func TestConsumerRebalance(t *testing.T) {

	var Topic = "test-rebalance-" + strconv.Itoa(int(time.Now().Unix()))

	createTopic(t, Topic, 2, 1)
	defer func() { removeTopic(t, Topic) }()

	onError := func(_ context.Context, err error) {
		require.NoError(t, err)
	}

	chMsg := make(chan *kafka.Message, 1)
	onProcess := func(_ context.Context, msg *kafka.Message) error {
		if msg == nil {
			return errors.New("invalid message")
		}
		chMsg <- msg
		return nil
	}

	onCommit := func(_ context.Context, topic string, partition int32, offset kafka.Offset) {
		require.Equal(t, Topic, topic)
		require.True(t, partition >= 0)
		require.True(t, offset >= 0)
	}

	c1 := newConsumer(t, Topic, nil, onError, onProcess, onCommit)
	defer func() { require.NoError(t, c1.Stop()) }()
	go func() { require.NoError(t, c1.Start()) }()

	c2 := newConsumer(t, Topic, nil, onError, onProcess, onCommit)
	defer func() { require.NoError(t, c2.Stop()) }()
	go func() { require.NoError(t, c2.Start()) }()

	time.Sleep(time.Second)
	require.NoError(t, c2.Stop())

	time.Sleep(time.Second * 4) // wait rebalance
}

func TestConsumerFailedSubscribe(t *testing.T) {

	onError := func(_ context.Context, err error) {
		require.NoError(t, err)
	}

	onProcess := func(_ context.Context, msg *kafka.Message) error {
		if msg == nil {
			return errors.New("invalid message")
		}
		return nil
	}

	onCommit := func(_ context.Context, topic string, partition int32, offset kafka.Offset) {
		require.NotEmpty(t, topic)
		require.True(t, partition >= 0)
		require.True(t, offset >= 0)
	}

	c1 := newConsumer(t, "", nil, onError, onProcess, onCommit)
	defer func() { require.NoError(t, c1.Stop()) }()

	require.EqualError(t, c1.Start(), "subscribe to topics failed: Local: Invalid argument or configuration")
}

func TestConsumerRevokePartition(t *testing.T) {

	var Topic = "test-revoke-" + strconv.Itoa(int(time.Now().Unix()))

	createTopic(t, Topic, 2, 1)
	defer func() { removeTopic(t, Topic) }()

	chErrors := make(chan error, 2)
	onError := func(_ context.Context, err error) {
		chErrors <- err
	}

	onProcess := func(_ context.Context, msg *kafka.Message) error {
		if msg == nil {
			return errors.New("invalid message")
		}
		return nil
	}

	onCommit := func(_ context.Context, topic string, partition int32, offset kafka.Offset) {
		require.Equal(t, Topic, topic)
		require.True(t, partition >= 0)
		require.True(t, offset >= 0)
	}

	c1 := newConsumer(t, Topic, nil, onError, onProcess, onCommit)
	defer func() { require.NoError(t, c1.Stop()) }()

	go func() { require.NoError(t, c1.Start()) }()

	c2 := newConsumer(t, Topic, nil, onError, onProcess, onCommit)
	defer func() { require.NoError(t, c2.Stop()) }()

	go func() { require.NoError(t, c2.Start()) }()

	time.Sleep(time.Second * 4) // wait rebalance

	removeTopic(t, Topic)

	time.Sleep(time.Second * 4) // wait revoke
}

func TestConsumerCommitToMonoTopic(t *testing.T) {

	const CountMessages = 100
	var Topic = "test-commit-mono-" + strconv.Itoa(int(time.Now().Unix()))

	createTopic(t, Topic, 1, 1)
	defer func() { removeTopic(t, Topic) }()

	onError := func(_ context.Context, err error) {
		zapLogger := newLogger(t)
		zapLogger.Fatal("failed to invoke", zap.Error(err))
	}

	chMsg := make(chan *kafka.Message, CountMessages)
	onProcess := func(_ context.Context, msg *kafka.Message) error {
		if msg == nil {
			return errors.New("invalid message")
		}
		chMsg <- msg
		return nil
	}

	onCommit := func(_ context.Context, topic string, partition int32, offset kafka.Offset) {
		require.Equal(t, Topic, topic)
		require.True(t, partition >= 0)
		require.True(t, offset >= 0)
	}

	props := kafka.ConfigMap{}

	fnNewConsumer := func() *Consumer {
		return newConsumer(t, Topic, props, onError, onProcess, onCommit)
	}

	for _, turnOnRebalance := range []bool{true} {
		for countConsumers := 1; countConsumers < 2; countConsumers++ {
			if t.Failed() {
				return
			}

			invokeCommitTest(t, turnOnRebalance, countConsumers, CountMessages, Topic, chMsg, fnNewConsumer)
		}
	}
}

func TestConsumerCommitToTopicWithMultiParts(t *testing.T) {

	const CountMessages = 151
	var Topic = "test-commit-multiparts-" + strconv.Itoa(int(time.Now().Unix()))

	createTopic(t, Topic, 5, 1)
	defer func() { removeTopic(t, Topic) }()

	onError := func(_ context.Context, err error) {
		zapLogger := newLogger(t)
		zapLogger.Fatal("failed to invoke", zap.Error(err))
	}

	chMsg := make(chan *kafka.Message, CountMessages)
	onProcess := func(_ context.Context, msg *kafka.Message) error {
		if msg == nil {
			return errors.New("invalid message")
		}
		chMsg <- msg
		return nil
	}

	onCommit := func(_ context.Context, topic string, partition int32, offset kafka.Offset) {
		require.Equal(t, Topic, topic)
		require.True(t, partition >= 0)
		require.True(t, offset >= 0)
	}

	props := kafka.ConfigMap{
		// Enable generation of PartitionEOF when the
		// end of a partition is reached.
		"enable.partition.eof": true,
	}

	fnNewConsumer := func() *Consumer {
		return newConsumer(t, Topic, props, onError, onProcess, onCommit)
	}

	for _, turnOnRebalance := range []bool{false, true} {
		for countConsumers := 1; countConsumers < 5; countConsumers++ {
			if t.Failed() {
				return
			}

			invokeCommitTest(t, turnOnRebalance, countConsumers, CountMessages, Topic, chMsg, fnNewConsumer)
		}
	}
}

func invokeCommitTest(t *testing.T, turnOnRebalance bool, countConsumers, countMessages int, topic string, chMsg <-chan *kafka.Message, createConsumer func() *Consumer) {
	t.Helper()

	testInfo := fmt.Sprintf("count consumers: %d, with rebalance: %v", countConsumers, turnOnRebalance)
	zapLogger := newLogger(t).With(
		zap.Int("count consumers", countConsumers),
		zap.Bool("with rebalance", turnOnRebalance))

	consumersList := make([]*Consumer, 0, countConsumers)
	defer func() {
		for i := range consumersList {
			require.NoError(t, consumersList[i].Stop(), testInfo)
		}
	}()

	for i := 0; i < countConsumers; i++ {
		c := createConsumer()
		go func() { require.NoError(t, c.Start(), testInfo) }()
		consumersList = append(consumersList, c)
	}

	p := newProducer(t, topic)
	defer p.Close()

	go func() {
		for i := 0; i < countMessages; i++ {
			deliveryChan := make(chan kafka.Event)
			require.NoError(t, p.Produce(
				&kafka.Message{
					TopicPartition: kafka.TopicPartition{
						Topic:     &topic,
						Partition: kafka.PartitionAny,
					},
					Value: []byte(strconv.Itoa(i)),
				},
				deliveryChan),
				testInfo)

			event := <-deliveryChan
			eventMessage, ok := event.(*kafka.Message)
			require.True(t, ok, testInfo+" %#v", event)
			require.NoError(t, eventMessage.TopicPartition.Error, testInfo+" %#v", event)
		}
	}()

	var (
		count, prevPercent int
		inMessages         = make([]int, countMessages)
	)
	for i := 0; i < len(inMessages); i++ {
		inMessages[i] = -1
	}

	for msg := range chMsg {
		msgID, err := strconv.Atoi(string(msg.Value))
		require.NoError(t, err, testInfo)
		if inMessages[msgID] > -1 {
			zapLogger.Fatal("not unique value", zap.Int("index", msgID), zap.Int("value", inMessages[msgID]))
		}
		inMessages[msgID] = msgID

		count++
		if count == countMessages {
			break
		}

		percent := int(float64(count) / float64(countMessages) * 100)

		if turnOnRebalance {
			if prevPercent != percent && percent%10 == 1 {
				zapLogger.Info("stop consumer")
				require.NoError(t, consumersList[0].Stop(), testInfo)

				zapLogger.Info("start consumer")
				consumersList[0] = createConsumer()
				go func() { require.NoError(t, consumersList[0].Start(), testInfo) }()
				time.Sleep(time.Second) // wait rebalance

			} else if countConsumers > 1 && countMessages > 10 && count == countMessages-10 {
				zapLogger.Info("stop consumer")
				require.NoError(t, consumersList[0].Stop(), testInfo)
			}
		}

		prevPercent = percent
	}

	select {
	case m := <-chMsg:
		zapLogger.Fatal("order is not empty", zap.String("value", string(m.Value)))
	default:
		// ok
	}

	for i := range inMessages {
		require.Equal(t, i, inMessages[i], testInfo)
	}
}

func newConsumer(t *testing.T, topic string, props kafka.ConfigMap, onError FuncOnError, onProcess FuncOnProcess, onCommit FuncOnCommit) *Consumer {

	zapLogger := newLogger(t)

	cfg := &Config{
		OnError:           onError,
		OnProcess:         onProcess,
		OnCommit:          onCommit,
		Topics:            []string{topic},
		CommitOffsetCount: 1,
		ConfigMap: &kafka.ConfigMap{
			"group.id":           "group-id",
			"bootstrap.servers":  getKafkaServers(),
			"auto.offset.reset":  "earliest",
			"session.timeout.ms": 6000,
		},
	}

	if len(props) > 0 {
		for k, v := range props {
			(*cfg.ConfigMap)[k] = v
		}
	}

	c, err := New(cfg, zapLogger)
	require.NoError(t, err)

	return c
}

func newProducer(t *testing.T, topic string) *kafka.Producer {

	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": getKafkaServers(),
	})
	require.NoError(t, err)

	return p
}

func newLogger(t *testing.T) *zap.Logger {

	zapCfg := zap.NewProductionConfig()
	zapCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapLogger, err := zapCfg.Build()
	require.NoError(t, err)

	return zapLogger
}

func removeTopic(t *testing.T, topic string) {

	c := newAdminClient(t)
	results, err := c.DeleteTopics(context.Background(), []string{topic})
	require.NoError(t, err)

	for _, res := range results {
		require.Equal(t, kafka.ErrNoError, res.Error.Code(), res.Error.String())
	}
}

func createTopic(t *testing.T, topic string, numParts, replicationFactor int) {

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

	c, err := kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers": getKafkaServers(),
	})
	require.NoError(t, err)

	return c
}

func getKafkaServers() string {
	return "localhost:9092"
}
