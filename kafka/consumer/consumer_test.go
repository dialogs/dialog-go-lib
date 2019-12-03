package consumer

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	log.SetFlags(log.Ltime | log.Lmicroseconds | log.Llongfile)
}

func TestConsumerNew(t *testing.T) {

	cfg := &Config{
		CommitOffsetCount:    1,
		CommitOffsetDuration: time.Hour,
		OnError:              func(context.Context, *zap.Logger, error) {},
		OnProcess:            func(context.Context, *zap.Logger, *kafka.Message, DelayI) error { return nil },
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

	onError := func(_ context.Context, _ *zap.Logger, err error) {
		require.NoError(t, err)
	}

	onProcess := func(_ context.Context, _ *zap.Logger, msg *kafka.Message, c DelayI) error {
		if msg == nil {
			return errors.New("invalid message")
		}
		return nil
	}

	onCommit := func(_ context.Context, _ *zap.Logger, topic string, partition int32, offset kafka.Offset, count int) {
		require.Equal(t, Topic, topic)
		require.True(t, partition >= 0)
		require.True(t, offset >= 0)
	}

	c := newConsumer(t, []string{Topic}, nil, onError, onProcess, onCommit, nil, nil, 0)
	go func() { require.NoError(t, c.Start()) }()
	defer c.Stop()

	time.Sleep(time.Second)
	c.Stop()

	require.EqualError(t, c.Start(), "consumer already closed")
}

func TestConsumerReadMessageSuccess(t *testing.T) {

	var Topic = "test-read-message-" + strconv.Itoa(int(time.Now().Unix()))

	createTopic(t, Topic, 1, 1)
	defer func() { removeTopic(t, Topic) }()

	onError := func(_ context.Context, _ *zap.Logger, err error) {
		require.NoError(t, err)
	}

	chMsg := make(chan *kafka.Message, 2)
	onProcess := func(_ context.Context, _ *zap.Logger, msg *kafka.Message, d DelayI) error {
		if msg == nil {
			return errors.New("invalid message")
		}
		chMsg <- msg
		return nil
	}

	onCommit := func(_ context.Context, _ *zap.Logger, topic string, partition int32, offset kafka.Offset, count int) {
		require.Equal(t, Topic, topic)
		require.True(t, partition >= 0)
		require.True(t, offset >= 0)
		require.Equal(t, 1, count)
	}

	c1 := newConsumer(t, []string{Topic}, nil, onError, onProcess, onCommit, nil, nil, 0)
	defer c1.Stop()

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
}

func TestConsumerReadMessagesWithDelay(t *testing.T) {

	var Topic = "test-read-message-" + strconv.Itoa(int(time.Now().Unix()))

	createTopic(t, Topic, 1, 1)
	defer func() { removeTopic(t, Topic) }()

	onError := func(_ context.Context, _ *zap.Logger, err error) {
		require.NoError(t, err)
	}

	chMsg := make(chan *kafka.Message, 2)
	delayDuration := 10 * time.Second
	onProcess := func(_ context.Context, _ *zap.Logger, msg *kafka.Message, d DelayI) error {
		if msg == nil {
			return errors.New("invalid message")
		}
		go func(t *testing.T, d DelayI) {
			require.NoError(t, d.DelayConsumer())
		}(t, d)
		chMsg <- msg
		return nil
	}

	onCommit := func(_ context.Context, _ *zap.Logger, topic string, partition int32, offset kafka.Offset, count int) {
		require.Equal(t, Topic, topic)
		require.True(t, partition >= 0)
		require.True(t, offset >= 0)
	}

	c1 := newConsumer(t, []string{Topic}, nil, onError, onProcess, onCommit, nil, nil, delayDuration)
	defer c1.Stop()

	go func() { require.NoError(t, c1.Start()) }()

	p := newProducer(t, Topic)
	defer p.Close()

	produceMessage := func(t *testing.T) {
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
	}

	produceMessage(t)

	c1.logger.Info("wait message")
	res := <-chMsg

	timeFirstMessage := time.Now()

	produceMessage(t)

	select {
	case <-c1.ctx.Done():
		require.True(t, false, "service shouldn't be closed")
	case <-chMsg:
		//continue
	}
	diff := time.Since(timeFirstMessage)
	require.True(t, diff > delayDuration, "diff = %v\n", diff)
	eps := 2 * time.Second
	require.True(t, diff < delayDuration+eps, "diff = %v\n", diff)
	c1.logger.Info("read message")

	require.Equal(t, &Topic, res.TopicPartition.Topic)
	require.Equal(t, int32(0), res.TopicPartition.Partition)
	require.Equal(t, []byte(Topic), res.Value)
}

func TestConsumerRebalance(t *testing.T) {

	const CountPartitions = 3
	var Topic = "test-rebalance-" + strconv.Itoa(int(time.Now().Unix()))

	createTopic(t, Topic, CountPartitions, 1)
	defer func() { removeTopic(t, Topic) }()

	onError := func(_ context.Context, _ *zap.Logger, err error) {
		require.NoError(t, err)
	}

	chMsg := make(chan *kafka.Message)
	onProcess := func(_ context.Context, _ *zap.Logger, msg *kafka.Message, c DelayI) error {
		if msg == nil {
			return errors.New("invalid message")
		}
		chMsg <- msg
		return nil
	}

	onCommit := func(_ context.Context, _ *zap.Logger, topic string, partition int32, offset kafka.Offset, count int) {
		require.Equal(t, Topic, topic)
		require.True(t, partition >= 0)
		require.True(t, offset >= 0)
	}

	chRebalance := make(chan int, 10)
	onRebalance := func(_ context.Context, _ *zap.Logger, topics []kafka.TopicPartition) {
		require.NotNil(t, topics)
		chRebalance <- len(topics)
	}

	c1 := newConsumer(t, []string{Topic}, nil, onError, onProcess, onCommit, onRebalance, onRebalance, 0)
	defer c1.Stop()
	go func() { require.NoError(t, c1.Start()) }()

	c2 := newConsumer(t, []string{Topic}, nil, onError, onProcess, onCommit, onRebalance, onRebalance, 0)
	defer c2.Stop()
	go func() { require.NoError(t, c2.Start()) }()

	waitRebalance(chRebalance)

	checkAssignmentBefore(t, c1, c2, CountPartitions)

	c2.Stop()

	waitRebalance(chRebalance)

	c1PartitionsAfter, err := c1.reader.Assignment()
	require.NoError(t, err)
	require.NotEmpty(t, c1PartitionsAfter)
	require.Len(t, c1PartitionsAfter, CountPartitions)
}

func TestConsumerFailedSubscribe(t *testing.T) {

	onError := func(_ context.Context, _ *zap.Logger, err error) {
		require.NoError(t, err)
	}

	onProcess := func(_ context.Context, _ *zap.Logger, msg *kafka.Message, c DelayI) error {
		if msg == nil {
			return errors.New("invalid message")
		}
		return nil
	}

	onCommit := func(_ context.Context, _ *zap.Logger, topic string, partition int32, offset kafka.Offset, count int) {
		require.NotEmpty(t, topic)
		require.True(t, partition >= 0)
		require.True(t, offset >= 0)
	}

	c1 := newConsumer(t, []string{""}, nil, onError, onProcess, onCommit, nil, nil, 0)
	defer c1.Stop()

	require.EqualError(t, c1.Start(), "subscribe to topics failed: Local: Invalid argument or configuration")
}

func TestConsumerRevokePartition(t *testing.T) {

	const CountPartitions = 3

	var Topic = "test-revoke-" + strconv.Itoa(int(time.Now().Unix()))

	createTopic(t, Topic, CountPartitions, 1)
	defer func() { removeTopic(t, Topic) }()

	chErrors := make(chan error, 2)
	onError := func(_ context.Context, _ *zap.Logger, err error) {
		chErrors <- err
	}

	onProcess := func(_ context.Context, _ *zap.Logger, msg *kafka.Message, c DelayI) error {
		if msg == nil {
			return errors.New("invalid message")
		}
		return nil
	}

	onCommit := func(_ context.Context, _ *zap.Logger, topic string, partition int32, offset kafka.Offset, count int) {
		require.Equal(t, Topic, topic)
		require.True(t, partition >= 0)
		require.True(t, offset >= 0)
	}

	chRebalance := make(chan int)
	onRebalance := func(_ context.Context, _ *zap.Logger, topics []kafka.TopicPartition) {
		require.NotNil(t, topics)
		chRebalance <- len(topics)
	}

	c1 := newConsumer(t, []string{Topic}, nil, onError, onProcess, onCommit, onRebalance, onRebalance, 0)
	defer c1.Stop()

	go func() { require.NoError(t, c1.Start()) }()

	c2 := newConsumer(t, []string{Topic}, nil, onError, onProcess, onCommit, onRebalance, onRebalance, 0)
	defer c2.Stop()

	go func() { require.NoError(t, c2.Start()) }()

	waitRebalance(chRebalance)

	checkAssignmentBefore(t, c1, c2, CountPartitions)

	removeTopic(t, Topic)

	waitRebalance(chRebalance)

	c1PartitionsAfter, err := c1.reader.Assignment()
	require.NoError(t, err)

	c2PartitionsAfter, err := c2.reader.Assignment()
	require.NoError(t, err)

	// WARN! Assignment return not empty partitions list
	require.Len(t, append(c1PartitionsAfter, c2PartitionsAfter...), 1)
}

func TestConsumerCommit(t *testing.T) {

	fnNewTopics := func(countTopic, countPartitions int) []string {
		topicList := make([]string, 0, countTopic)
		for i := 0; i < countTopic; i++ {
			topic := fmt.Sprintf("test-commit-%d-%d-%d", i, countPartitions, int(time.Now().Unix()))

			createTopic(t, topic, countPartitions, 1)
			topicList = append(topicList, topic)
		}

		return topicList
	}

	fnNewConsumer := func(topicList []string, countMessages int) (func() *Consumer, chan *kafka.Message, chan int) {

		props := kafka.ConfigMap{
			// Enable generation of PartitionEOF when the
			// end of a partition is reached.
			"enable.partition.eof": true,
		}

		onError := func(_ context.Context, _ *zap.Logger, err error) {
			panic(err)
		}

		chMsg := make(chan *kafka.Message, countMessages)
		onProcess := func(_ context.Context, _ *zap.Logger, msg *kafka.Message, c DelayI) error {
			if msg == nil {
				return errors.New("invalid message")
			}
			chMsg <- msg
			return nil
		}

		onCommit := func(_ context.Context, _ *zap.Logger, topic string, partition int32, offset kafka.Offset, count int) {
			require.Contains(t, topicList, topic)
			require.True(t, partition >= 0)
			require.True(t, offset >= 0)
		}

		chRebalance := make(chan int, 10000000) // buffer: channel use only in a start of a test (waiting of consumers)
		onRebalance := func(_ context.Context, _ *zap.Logger, topics []kafka.TopicPartition) {
			chRebalance <- len(topics)
		}

		return func() *Consumer {
				return newConsumer(t, topicList, props, onError, onProcess, onCommit, onRebalance, onRebalance, 0)
			},
			chMsg,
			chRebalance
	}

	for _, testInfo := range []struct {
		Topics        int
		Partitions    int
		Consumers     int
		Rebalance     bool
		CountMessages int
	}{
		{Topics: 1, Partitions: 1, Consumers: 1, Rebalance: false, CountMessages: 30},
		{Topics: 1, Partitions: 1, Consumers: 1, Rebalance: true, CountMessages: 30},
		{Topics: 1, Partitions: 3, Consumers: 4, Rebalance: true, CountMessages: 30},
		{Topics: 2, Partitions: 3, Consumers: 4, Rebalance: true, CountMessages: 11},
		{Topics: 2, Partitions: 10, Consumers: 40, Rebalance: false, CountMessages: 200},
		{Topics: 2, Partitions: 10, Consumers: 40, Rebalance: true, CountMessages: 200},
	} {

		name := fmt.Sprintf("topics: %d; partitions: %d; consumers: %d; rebalance: %v; messages: %d",
			testInfo.Topics,
			testInfo.Partitions,
			testInfo.Consumers,
			testInfo.Rebalance,
			testInfo.CountMessages)

		fn := func(*testing.T) {
			topicList := fnNewTopics(testInfo.Topics, testInfo.Partitions)
			defer func() { removeTopic(t, topicList...) }()

			fnConsumer, chMsg, chRebalance := fnNewConsumer(topicList, testInfo.CountMessages)
			defer close(chRebalance)
			defer close(chMsg)

			invokeCommitTest(t, testInfo.Rebalance, testInfo.Consumers, testInfo.CountMessages, topicList, chMsg, chRebalance, fnConsumer)
		}

		if !t.Run(name, fn) {
			return
		}
	}
}

func invokeCommitTest(t *testing.T, turnOnRebalance bool, countConsumers, countMessages int, topicList []string, chMsg <-chan *kafka.Message, chRebalance <-chan int, createConsumer func() *Consumer) {
	t.Helper()

	testInfo := fmt.Sprintf("count consumers: %d, with rebalance: %v", countConsumers, turnOnRebalance)
	zapLogger := newLogger(t).With(
		zap.Int("count consumers", countConsumers),
		zap.Bool("with rebalance", turnOnRebalance))

	consumersList := make([]*Consumer, 0, countConsumers)
	defer func() {
		for i := range consumersList {
			consumersList[i].Stop()
		}
	}()

	for i := 0; i < countConsumers; i++ {
		c := createConsumer()
		go func() { require.NoError(t, c.Start(), testInfo) }()
		consumersList = append(consumersList, c)
	}

	waitRebalance(chRebalance)

	for _, topicName := range topicList {
		go func(topic string) {
			p := newProducer(t, topic)
			defer p.Close()

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
		}(topicName)
	}

	var (
		count, prevPercent int
		inMessages         = make(map[string][]int)
	)

	for _, topic := range topicList {
		value := make([]int, countMessages)
		for i := 0; i < len(value); i++ {
			value[i] = -1
		}

		inMessages[topic] = value
	}

	for msg := range chMsg {
		msgID, err := strconv.Atoi(string(msg.Value))
		require.NoError(t, err, testInfo)

		arr := inMessages[*msg.TopicPartition.Topic]
		if arr[msgID] > -1 {
			zapLogger.Fatal("not unique value", zap.Int("index", msgID), zap.Int("value", arr[msgID]))
		}
		arr[msgID] = msgID

		count++
		if count == (countMessages * len(topicList)) {
			break
		}

		percent := int(float64(count) / float64(countMessages) * 100)

		if turnOnRebalance {
			if prevPercent != percent && percent%30 == 1 {
				zapLogger.Info("stop consumer")
				consumersList[0].Stop()

				time.Sleep(time.Second * 4) // wait rebalance

				zapLogger.Info("start consumer")
				consumersList[0] = createConsumer()
				go func() { require.NoError(t, consumersList[0].Start(), testInfo) }()

				time.Sleep(time.Second * 4) // wait rebalance

			} else if countConsumers > 1 && countMessages > 10 && count == countMessages-10 {
				zapLogger.Info("stop consumer")
				consumersList[0].Stop()

				time.Sleep(time.Second * 4) // wait rebalance
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

	require.Equal(t, len(topicList), len(inMessages))
	for _, topic := range topicList {
		arr := inMessages[topic]
		for i := range arr {
			require.Equal(t, i, arr[i], testInfo)
		}
	}
}

func newConsumer(t *testing.T, topicList []string, props kafka.ConfigMap, onError FuncOnError, onProcess FuncOnProcess, onCommit FuncOnCommit, onRevoke FuncOnRevoke, onRebalance FuncOnRebalance, delay time.Duration) *Consumer {
	t.Helper()

	zapLogger := newLogger(t)

	cfg := newConsumerConfig(topicList, props, onError, onProcess, onCommit, onRevoke, onRebalance, delay)

	c, err := New(cfg, zapLogger)
	require.NoError(t, err)

	return c
}

func newConsumerConfig(topicList []string, props kafka.ConfigMap, onError FuncOnError, onProcess FuncOnProcess, onCommit FuncOnCommit, onRevoke FuncOnRevoke, onRebalance FuncOnRebalance, delay time.Duration) *Config {

	cfg := &Config{
		OnError:           onError,
		OnProcess:         onProcess,
		OnCommit:          onCommit,
		OnRevoke:          onRevoke,
		OnRebalance:       onRebalance,
		Topics:            topicList,
		CommitOffsetCount: 11,
		ConfigMap: &kafka.ConfigMap{
			"group.id":           "group-id",
			"bootstrap.servers":  getKafkaServers(),
			"auto.offset.reset":  "earliest",
			"session.timeout.ms": 6000,
		},
		Delay: delay,
	}

	if len(props) > 0 {
		for k, v := range props {
			(*cfg.ConfigMap)[k] = v
		}
	}

	return cfg
}

func newProducer(t *testing.T, topic string) *kafka.Producer {
	t.Helper()

	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": getKafkaServers(),
	})
	require.NoError(t, err)

	return p
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

func checkAssignmentBefore(t *testing.T, c1, c2 *Consumer, countPartitions int) {
	t.Helper()

	c1PartitionsBefore, err := c1.reader.Assignment()
	require.NoError(t, err)

	c2PartitionsBefore, err := c2.reader.Assignment()
	require.NoError(t, err)

	require.NotContains(t, c1PartitionsBefore, c2PartitionsBefore)
	require.NotContains(t, c2PartitionsBefore, c1PartitionsBefore)
	require.Len(t, append(c1PartitionsBefore, c2PartitionsBefore...), countPartitions)
}

func waitRebalance(chRebalance <-chan int) {

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		total := 0

		for {
			select {
			case <-time.After(time.Second * 5):
				if total > 0 {
					return
				}
			case count := <-chRebalance:
				total += count
			}
		}
	}()

	wg.Wait()
}
