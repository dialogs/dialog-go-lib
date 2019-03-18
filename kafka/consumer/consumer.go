package consumer

import (
	"container/heap"
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

var nopCommitFunc = func(ctx context.Context, partition int32, offset kafka.Offset, committed int) {}

type Config struct {
	OnCommit     func(ctx context.Context, partition int32, offset kafka.Offset, committed int)
	OnError      func(ctx context.Context, err error)
	OnProcess    func(ctx context.Context, msg *kafka.Message)
	PoolTimeout  time.Duration
	QueueSize    int
	ReaderConfig *kafka.ConfigMap
	Topics       []string
	WorkersCount int
}

type Consumer struct {
	cancel       context.CancelFunc
	ctx          context.Context
	commitQueue  chan *kafka.Message
	onCommit     func(ctx context.Context, partition int32, offset kafka.Offset, committed int)
	onError      func(ctx context.Context, err error)
	onProcess    func(ctx context.Context, msg *kafka.Message)
	poolTimeout  time.Duration
	processQueue chan *kafka.Message
	reader       *kafka.Consumer
	rebalancer   chan *Partition
	topics       []string
	workersCount int
	wg           sync.WaitGroup
}

type Partition struct {
	num    int32
	offset kafka.Offset
}

func NewConfig() *Config {
	return &Config{
		PoolTimeout:  1 * time.Second,
		QueueSize:    10,
		WorkersCount: 10,
	}
}

func New(cfg *Config) (*Consumer, error) {
	if cfg.QueueSize == 0 {
		return nil, errors.New("consumer group queue size is 0")
	}

	if cfg.WorkersCount == 0 {
		return nil, errors.New("consumer group queue size is 0")
	}

	if cfg.OnError == nil {
		return nil, errors.New("on error callback is nil")
	}

	if cfg.OnProcess == nil {
		return nil, errors.New("on process callback is nil")
	}

	if len(cfg.Topics) == 0 {
		return nil, errors.New("topics is empty")
	}

	if cfg.ReaderConfig == nil {
		return nil, errors.New("reader config is nil")
	}

	var onCommit func(ctx context.Context, partition int32, offset kafka.Offset, committed int)
	if cfg.OnCommit != nil {
		onCommit = cfg.OnCommit
	} else {
		onCommit = nopCommitFunc
	}

	ctx, cancel := context.WithCancel(context.Background())

	reader, err := kafka.NewConsumer(cfg.ReaderConfig)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		cancel:       cancel,
		ctx:          ctx,
		commitQueue:  make(chan *kafka.Message, cfg.QueueSize),
		onCommit:     onCommit,
		onError:      cfg.OnError,
		onProcess:    cfg.OnProcess,
		poolTimeout:  cfg.PoolTimeout,
		processQueue: make(chan *kafka.Message, cfg.QueueSize),
		reader:       reader,
		rebalancer:   make(chan *Partition),
		topics:       cfg.Topics,
		workersCount: cfg.WorkersCount,
	}, nil
}

func (c *Consumer) Start() error {
	err := c.reader.SubscribeTopics(c.topics, c.rebalance)
	if err != nil {
		c.reader.Close()
		return err
	}

	c.wg.Add(1)
	go func() {
		c.readLoop()
		c.wg.Done()
	}()

	c.wg.Add(1)
	go func() {
		c.commitLoop()
		c.wg.Done()
	}()

	c.wg.Add(c.workersCount)
	for i := 0; i < c.workersCount; i++ {
		go func() {
			c.processLoop()
			c.wg.Done()
		}()
	}

	return nil
}

func (c *Consumer) Stop() {
	c.cancel()
	c.reader.Close()
	c.wg.Wait()
}

func (c *Consumer) commitLoop() {
	targets := make([]*kafka.Message, 0)
	partitions := make(map[int32]*MessageHeap)
	offsets := make(map[int32]kafka.Offset)

	for {
		select {
		case <-c.ctx.Done():
			return
		case partition := <-c.rebalancer:
			_, ok := partitions[partition.num]
			if ok {
				partitions[partition.num] = new(MessageHeap)
			}

			offsets[partition.num] = partition.offset
		case msg := <-c.commitQueue:
			expectedOffset := offsets[msg.TopicPartition.Partition]

			if msg.TopicPartition.Offset < expectedOffset {
				continue
			}

			h, ok := partitions[msg.TopicPartition.Partition]
			if !ok {
				h = new(MessageHeap)
				partitions[msg.TopicPartition.Partition] = h
			}

			heap.Push(h, msg)

			// extract longest increasing subsequence starting from expected offset
			for h.Len() != 0 && (*h)[0].TopicPartition.Offset == expectedOffset {
				expectedOffset++
				offsets[msg.TopicPartition.Partition] = expectedOffset
				targets = append(targets, heap.Pop(h).(*kafka.Message))
			}
			if len(targets) > 0 {
				if _, err := c.reader.CommitMessage(targets[len(targets)-1]); err != nil {
					if err == context.Canceled {
						return
					}
					c.onError(c.ctx, err)
					continue
				}

				c.onCommit(c.ctx, msg.TopicPartition.Partition, expectedOffset, len(targets))

				targets = targets[:0]
			}
		}
	}
}

func (c *Consumer) rebalance(consumer *kafka.Consumer, event kafka.Event) error {
	switch e := event.(type) {
	case kafka.AssignedPartitions:
		err := c.reader.Assign(e.Partitions)
		if err != nil {
			c.onError(c.ctx, err)
			return err
		}
	case kafka.RevokedPartitions:
		err := c.reader.Unassign()
		if err != nil {
			c.onError(c.ctx, err)
			return err
		}
	case kafka.Error:
		c.onError(c.ctx, e)
	}

	return nil
}

func (c *Consumer) commitMessage(msg *kafka.Message) {
	select {
	case <-c.ctx.Done():
		return
	case c.commitQueue <- msg:

	}
}

func (c *Consumer) processLoop() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case msg := <-c.processQueue:
			c.onProcess(c.ctx, msg)
			c.commitMessage(msg)
		}
	}
}

func (c *Consumer) readLoop() {
	offsets := make(map[int32]kafka.Offset)

	for {
		msg, ok := c.readNext()
		if !ok {
			return
		}

		offset, ok := offsets[msg.TopicPartition.Partition]
		if !ok || offset+1 != msg.TopicPartition.Offset {
			offsets[msg.TopicPartition.Partition] = msg.TopicPartition.Offset
			select {
			case <-c.ctx.Done():
				return
			case c.rebalancer <- &Partition{num: msg.TopicPartition.Partition, offset: msg.TopicPartition.Offset}:
			}
		} else {
			offsets[msg.TopicPartition.Partition]++
		}

		select {
		case <-c.ctx.Done():
		case c.processQueue <- msg:

		}
	}
}

func (c *Consumer) readNext() (*kafka.Message, bool) {
	for {
		msg, err := c.reader.ReadMessage(c.poolTimeout)
		if err != nil {
			if err == io.EOF {
				return nil, false
			}

			switch typedErr := err.(type) {
			case kafka.Error:
				if typedErr.Code() == kafka.ErrTimedOut {
					select {
					case <-c.ctx.Done():
					default:
						continue
					}
				}
			}
			c.onError(c.ctx, err)
		}

		return msg, true
	}
}
