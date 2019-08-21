package consumer

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type FuncOnError func(ctx context.Context, err error)
type FuncOnProcess func(ctx context.Context, msg *kafka.Message)
type FuncOnCommit func(ctx context.Context, partition int32, offset kafka.Offset, committed int)

var nopCommitFunc = func(ctx context.Context, partition int32, offset kafka.Offset, committed int) {}

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

	requiredProps := kafka.ConfigMap{
		// https://godoc.org/github.com/confluentinc/confluent-kafka-go/kafka#hdr-Consumer_events
		"enable.auto.commit": false,
		// https://godoc.org/github.com/confluentinc/confluent-kafka-go/kafka#NewConsumer
		"go.events.channel.enable": true,
		"go.events.channel.size":   0, // don't use channel buffer
		// https://godoc.org/github.com/confluentinc/confluent-kafka-go/kafka#NewConsumer
		"go.application.rebalance.enable": true,
	}
	for k, v := range requiredProps {
		if err := cfg.ConfigMap.SetKey(k, v); err != nil {
			return nil, errors.Wrapf(err, "force set config %s to %v failed", k, v)
		}
	}

	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
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

	err := c.reader.SubscribeTopics(c.topics, nil)
	if err != nil {
		defer func() {
			if err := c.Stop(); err != nil {
				c.logger.Error("failed to stop", zap.Error(err))
			}
		}()

		return errors.Wrap(err, "subscribe to topics failed")
	}

	c.wg.Add(1)
	go func() {
		defer func() {
			if err := c.Stop(); err != nil {
				c.logger.Error("failed to stop", zap.Error(err))
			}
		}()

		defer c.wg.Done()

		c.listen()
	}()

	return nil
}

func (c *Consumer) Stop() (err error) {

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
		defer func() {
			c.logger.Info("done")
		}()

		err = c.reader.Close()
	}

	return
}

func (c *Consumer) listen() {

	//consumerOffsetsTmp := newOffset() // TODO: remove!

	consumerOffsets := newOffset()
	defer func() {
		c.commitOffsets(consumerOffsets)
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
			return

		case <-offsetsTicker.C:
			c.commitOffsets(consumerOffsets)

		case ev := <-c.reader.Events():

			switch e := ev.(type) {
			case kafka.AssignedPartitions:
				// consumer group rebalance event: assigned partition set
				c.logger.Info("assigned", zap.String("partitions", topicPartitionsToString(e.Partitions)))
				for _, p := range e.Partitions {
					if p.Error != nil {
						c.logger.Error("assigned", zap.Error(p.Error))
					}
				}

				err := c.reader.Assign(e.Partitions)
				if err != nil {
					c.logger.Error("assigned", zap.Error(err))
					c.onError(c.ctx, err)
					return
				}

				// committedOffsets, err := c.reader.Committed(e.Partitions, 5000)
				// if err != nil {
				// 	c.logger.Error("assigned", zap.Error(err))
				// 	c.onError(c.ctx, err)
				// 	return
				// }
				committedOffsets := e.Partitions

				// c.commitOffsets(consumerOffsets)
				consumerOffsets.Sync(committedOffsets...)
				// consumerOffsetsTmp = newOffset()
				// for i := range committedOffsets {
				// 	if committedOffsets[i].Offset > 0 {
				// 		consumerOffsetsTmp.Add(committedOffsets[i])
				// 	}
				// }
				c.logger.Debug("===--------> offset", zap.Any("offsets info", committedOffsets), zap.Any("assigned data", e.Partitions))

			case kafka.RevokedPartitions:
				// consumer group rebalance event: revoked partition set
				c.logger.Info("revoked", zap.String("partitions", topicPartitionsToString(e.Partitions)))
				for _, p := range e.Partitions {
					if p.Error != nil {
						c.logger.Error("revoked", zap.Error(p.Error))
					}
				}

				err := c.reader.Unassign()
				if err != nil {
					c.onError(c.ctx, err)
					return
				}

				// c.commitOffsets(consumerOffsets)
				consumerOffsets.Sync(e.Partitions...)
				//consumerOffsetsTmp.Sync(e.Partitions...)
				log.Println("====> revoked:", e.Partitions)

			case *kafka.Message:
				if e.TopicPartition.Error != nil {
					c.logger.Error("message", zap.Error(e.TopicPartition.Error))
				}

				// off := consumerOffsetsTmp.GetOffset(e.TopicPartition.Topic, e.TopicPartition.Partition)
				// if off > 0 && e.TopicPartition.Offset <= off {
				// 	c.logger.Warn("skip message", zap.Any("topic", e.TopicPartition))
				// 	continue // skip, already processed. TODO: fix that!
				// }

				c.onProcess(c.ctx, e)

				consumerOffsets.Add(e.TopicPartition)

				if c.commitOffsetCount > 0 {
					if consumerOffsets.Counter() >= c.commitOffsetCount {
						log.Println("====> commit:", e.TopicPartition)
						c.commitOffsets(consumerOffsets)
					}
				}

			case kafka.PartitionEOF:
				// consumer reached end of partition
				// Needs to be explicitly enabled by setting the `enable.partition.eof`
				// configuration property to true.
				if e.Error != nil {
					if kafkaErr, ok := e.Error.(kafka.Error); !ok || kafkaErr.Code() != kafka.ErrPartitionEOF {
						c.logger.Error("reached", zap.Error(e.Error))
					}
				}

				var topic string
				if e.Topic != nil {
					topic = *e.Topic
				}

				c.logger.Info("reached",
					zap.String("topic", topic),
					zap.Int32("partition", e.Partition),
					zap.Int64("offset", int64(e.Offset)))

			case kafka.OffsetsCommitted:
				// reports committed offsets
				// https://godoc.org/github.com/confluentinc/confluent-kafka-go/kafka#hdr-Consumer_events
				// Offset commit results (when `enable.auto.commit` is enabled)

				if e.Error != nil {
					c.logger.Error("committed offsets", zap.Error(e.Error))
				}

				for _, p := range e.Offsets {
					if p.Error != nil {
						c.logger.Error("committed offsets", zap.Error(p.Error))
					}
				}

				c.logger.Info("committed", zap.String("partitions", topicPartitionsToString(e.Offsets)))

			case kafka.Error:
				// Errors should generally be considered as informational, the client will try to automatically recover
				c.logger.Error("error", zap.Error(e))
				c.onError(c.ctx, e)

			default:
				c.logger.Error("unknown event", zap.Any("data", e))
			}

		}
	}
}

func (c *Consumer) commitOffsets(consumerOffsets *offset) {

	for p := consumerOffsets.Get(); p != nil; p = consumerOffsets.Get() {
		_, err := c.reader.CommitOffsets([]kafka.TopicPartition{*p})
		if err != nil && err != context.Canceled {
			c.logger.Error("failed to commit offset", zap.Any("data", p), zap.Error(err))

			c.onError(c.ctx, err)
			continue
		}
		c.logger.Debug("success to commit offset", zap.Any("data", p))

		// committedOffsets, err := c.reader.Committed([]kafka.TopicPartition{{
		// 	Topic:     p.Topic,
		// 	Partition: p.Partition,
		// }}, 5000)
		// if err == nil {
		// 	c.logger.Debug("===>success to commit offset", zap.Any("data", committedOffsets))
		// }

		consumerOffsets.Remove(*p)
		c.onCommit(c.ctx, p.Partition, p.Offset, 0)
	}
}
