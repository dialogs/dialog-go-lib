package consumer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type FuncOnError func(ctx context.Context, err error)
type FuncOnProcess func(ctx context.Context, msg *kafka.Message) error
type FuncOnCommit func(ctx context.Context, topic string, partition int32, offset kafka.Offset)

var nopCommitFunc = func(ctx context.Context, topic string, partition int32, offset kafka.Offset) {}

type Consumer struct {
	id                   uuid.UUID
	commitOffsetCount    int
	commitOffsetDuration time.Duration
	ctx                  context.Context
	ctxCancel            context.CancelFunc
	logger               *zap.Logger
	onCommit             FuncOnCommit
	onError              FuncOnError
	onProcess            FuncOnProcess
	reader               *kafka.Consumer
	topics               []string
	wg                   sync.WaitGroup
}

func New(cfg *Config, logger *zap.Logger) (*Consumer, error) {

	if err := cfg.Check(); err != nil {
		return nil, err
	}

	var onCommit FuncOnCommit
	if cfg.OnCommit != nil {
		onCommit = cfg.OnCommit
	} else {
		onCommit = nopCommitFunc
	}

	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	requiredProps := kafka.ConfigMap{
		// https://godoc.org/github.com/confluentinc/confluent-kafka-go/kafka#hdr-Consumer_events
		"enable.auto.commit": false,
		// https://godoc.org/github.com/confluentinc/confluent-kafka-go/kafka#NewConsumer
		"go.events.channel.enable": true,
		"go.events.channel.size":   0, // don't use channel buffer
		// https://godoc.org/github.com/confluentinc/confluent-kafka-go/kafka#NewConsumer
		"go.application.rebalance.enable": true,
		// https://docs.confluent.io/3.3.1/clients/librdkafka/CONFIGURATION_8md.html
		"api.version.request": "true",
		"client.id":           id.String(),
	}
	for k, v := range requiredProps {
		if err := cfg.ConfigMap.SetKey(k, v); err != nil {
			return nil, errors.Wrapf(err, "force set config %s to %v failed", k, v)
		}
	}

	logger = logger.With(zap.String("consumer", id.String()))

	ctx, ctxCancel := context.WithCancel(context.Background())

	reader, err := kafka.NewConsumer(cfg.ConfigMap)
	if err != nil {
		defer ctxCancel()
		return nil, errors.Wrap(err, "create reader failed")
	}

	return &Consumer{
		id:                   id,
		ctx:                  ctx,
		ctxCancel:            ctxCancel,
		logger:               logger,
		onCommit:             onCommit,
		onError:              cfg.OnError,
		onProcess:            cfg.OnProcess,
		reader:               reader,
		topics:               cfg.Topics,
		commitOffsetCount:    cfg.CommitOffsetCount,
		commitOffsetDuration: cfg.CommitOffsetDuration,
	}, nil
}

func (c *Consumer) Start() error {

	c.logger.Info("start")

	select {
	case <-c.ctx.Done():
		err := errors.New("consumer already closed")
		c.logger.Error("failed to start consumer", zap.Error(err))
		return err
	default:
		// ok
	}

	defer func() {
		c.logger.Info("closing...")
		// logs for issues:
		// https://github.com/confluentinc/confluent-kafka-go/issues/65
		// https://github.com/confluentinc/confluent-kafka-go/issues/189
		defer c.logger.Info("success closed")

		select {
		case <-c.ctx.Done():
			// nothing do
		default:
			if err := c.Stop(); err != nil {
				c.logger.Error("failed to stop", zap.Error(err))
			}
		}
	}()

	err := c.reader.SubscribeTopics(c.topics, nil)
	if err != nil {
		return errors.Wrap(err, "subscribe to topics failed")
	}

	c.wg.Add(1)
	defer c.wg.Done()

	return c.listen()
}

func (c *Consumer) Stop() (err error) {

	defer func() {
		if err := recover(); err != nil {
			c.logger.Error("stop: panic:" + fmt.Sprintln(err))
		}
	}()

	// protection for error: 'fatal error: unexpected signal during runtime execution'
	var isClosed bool
	select {
	case <-c.ctx.Done():
		isClosed = true
	default:
		// nothing do
	}

	c.ctxCancel()
	c.wg.Wait()

	if !isClosed {
		defer c.logger.Info("done")
		err = c.reader.Close()
	}

	return
}

func (c *Consumer) listen() error {

	c.logger.Info("start listener")

	consumerOffsets := newOffset()
	defer func() {
		if err := recover(); err != nil {
			c.logger.Error("listener: panic:" + fmt.Sprintln(err))
		}

		c.commitOffsets(consumerOffsets)
		c.logger.Info("close listener")
	}()

	commitOffsetDuration := c.commitOffsetDuration
	if commitOffsetDuration <= 0 {
		commitOffsetDuration = time.Second * 5
	}

	offsetsTicker := time.NewTicker(commitOffsetDuration)
	defer offsetsTicker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return nil

		case <-offsetsTicker.C:
			c.commitOffsets(consumerOffsets)

		case ev := <-c.reader.Events():

			switch e := ev.(type) {
			case kafka.AssignedPartitions:
				// consumer group rebalance event: assigned partition set

				opLog := c.logger.With(zap.String("operation", "rebalance"), zap.Any("event", e))

				if err := checkPartitions(e.Partitions); err != nil {
					opLog.Error("failed to check assigned partitions", zap.Error(err))
					c.onError(c.ctx, err)
					return err
				}

				// Issue: https://github.com/confluentinc/confluent-kafka-go/issues/212
				// All offsets in topics are equal -1001(unset).
				// If assign the event with these invalid value, duplicate messages can appear.
				// Fix: read committed offsets before assign new topics.
				committedOffsets, err := c.reader.Committed(e.Partitions, 5000)
				if err != nil {
					opLog.Error("failed to read committed offsets", zap.Error(err))
					c.onError(c.ctx, err)
					return err
				}

				for i := range committedOffsets {
					item := &committedOffsets[i]
					if item.Offset >= 0 {
						item.Offset++
					}
				}

				if err := c.reader.Assign(committedOffsets); err != nil {
					opLog.Error("failed to set assigned", zap.Error(err))
					c.onError(c.ctx, err)
					return err
				}

				opLog.Info("success")

			case kafka.RevokedPartitions:
				// consumer group rebalance event: revoked partition set

				opLog := c.logger.With(zap.String("operation", "revoked"), zap.Any("event", e))

				if err := checkPartitions(e.Partitions); err != nil {
					opLog.Error("failed to check revoked partitions", zap.Error(err))
					c.onError(c.ctx, err)
					return err
				}

				err := c.reader.Unassign()
				if err != nil {
					opLog.Error("failed to unassign", zap.Error(err))
					c.onError(c.ctx, err)
					return err
				}

				consumerOffsets.Sync(e.Partitions...)

				opLog.Info("success")

			case *kafka.Message:

				opLog := c.logger.With(zap.String("operation", "message"), zap.Any("event", e))

				if err := e.TopicPartition.Error; err != nil {
					opLog.Error("failed to check", zap.Error(err))
					c.onError(c.ctx, err)
					return err
				}

				if err := c.onProcess(c.ctx, e); err != nil {
					opLog.Error("failed to process message", zap.Error(err))
					return err
				}

				consumerOffsets.Add(e.TopicPartition)

				if c.commitOffsetCount > 0 {
					if consumerOffsets.Counter() >= c.commitOffsetCount {
						c.commitOffsets(consumerOffsets)
					}
				}

				opLog.Debug("success")

			case kafka.PartitionEOF:
				// consumer reached end of partition
				// Needs to be explicitly enabled by setting the `enable.partition.eof`
				// configuration property to true.

				opLog := c.logger.With(zap.String("operation", "partition EOF"), zap.Any("event", e))

				if err := e.Error; err != nil {
					if kafkaErr, ok := err.(kafka.Error); !ok || kafkaErr.Code() != kafka.ErrPartitionEOF {
						opLog.Error("failed to check", zap.Error(err))
						c.onError(c.ctx, err)
						return err
					}
				}

				opLog.Info("success")

			case kafka.OffsetsCommitted:
				// reports committed offsets
				// https://godoc.org/github.com/confluentinc/confluent-kafka-go/kafka#hdr-Consumer_events
				// Offset commit results (when `enable.auto.commit` is enabled)

				opLog := c.logger.With(zap.String("operation", "committed offsets"), zap.Any("event", e))

				if err := e.Error; err != nil {
					opLog.Error("committed offsets", zap.Error(err))
					c.onError(c.ctx, err)
					return err
				}

				if err := checkPartitions(e.Offsets); err != nil {
					opLog.Error("failed to check", zap.Error(err))
					c.onError(c.ctx, err)
					return err
				}

				opLog.Info("success")

			case kafka.Error:
				// Errors should generally be considered as informational, the client will try to automatically recover
				c.logger.Error("error event", zap.Error(e))
				c.onError(c.ctx, e)

			default:
				c.logger.Error("unknown event", zap.Any("payload", e))
			}

		}
	}
}

func (c *Consumer) commitOffsets(consumerOffsets *offset) {

	list := consumerOffsets.Get()
	if len(list) > 0 {
		opLog := c.logger.WithOptions(zap.AddCallerSkip(1)).With(
			zap.String("operation", "commit offsets"),
			zap.Any("event", list))

		success, err := c.reader.CommitOffsets(list)
		if err != nil {
			opLog.Error("failed to commit", zap.Error(err))
			c.onError(c.ctx, err)
			return
		}

		if err := checkPartitions(success); err != nil {
			opLog.Error("failed to commit", zap.Error(err))
			c.onError(c.ctx, err)
			return
		}

		opLog.Debug("success", zap.Any("result", success))

		var topic string
		for i := range success {
			item := &success[i]
			consumerOffsets.Remove(*item)

			if item.Topic != nil {
				topic = *item.Topic
			}

			c.onCommit(c.ctx, topic, item.Partition, item.Offset)
		}
	}
}
