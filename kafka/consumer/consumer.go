package consumer

import (
	"container/heap"
	"context"
	"errors"
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

	reader, err := kafka.NewConsumer(cfg.ReaderConfig)
	if err != nil {
		return nil, err
	}

	err = reader.SubscribeTopics(cfg.Topics, nil)
	if err != nil {
		reader.Close()
		return nil, err
	}

	var onCommit func(ctx context.Context, partition int32, offset kafka.Offset, committed int)
	if cfg.OnCommit != nil {
		onCommit = cfg.OnCommit
	} else {
		onCommit = nopCommitFunc
	}

	ctx, cancel := context.WithCancel(context.Background())

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
		workersCount: cfg.WorkersCount,
	}, nil
}

func (c *Consumer) Start() {
	c.wg.Add(1)
	go func() {
		c.eventLoop()
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

}

func (c *Consumer) Stop() {
	c.cancel()
	c.wg.Wait()
	c.reader.Close()
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
			if partition != nil {
				offsets[partition.num] = partition.offset
			} else {
				offsets = make(map[int32]kafka.Offset)
				partitions = make(map[int32]*MessageHeap)
			}
		case msg := <-c.commitQueue:
			expectedOffset, ok := offsets[msg.TopicPartition.Partition]
			if !ok {
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

func (c *Consumer) commitMessage(msg *kafka.Message) {
	select {
	case <-c.ctx.Done():
		return
	case c.commitQueue <- msg:

	}
}

func (c *Consumer) eventLoop() {
	var offsets map[int32]kafka.Offset

	for {
		select {
		case <-c.ctx.Done():
			return
		case ev := <-c.reader.Events():
			switch e := ev.(type) {
			case kafka.AssignedPartitions:
				err := c.reader.Assign(e.Partitions)
				if err != nil {
					c.onError(c.ctx, err)
				}
				offsets = make(map[int32]kafka.Offset)
			case kafka.RevokedPartitions:
				err := c.reader.Unassign()
				if err != nil {
					c.onError(c.ctx, err)
				}
				offsets = nil
				select {
				case <-c.ctx.Done():
					return
				case c.rebalancer <- nil:
				}
			case *kafka.Message:
				if e.TopicPartition.Error != nil {
					continue
				}
				_, ok := offsets[e.TopicPartition.Partition]
				if !ok {
					select {
					case <-c.ctx.Done():
						return
					case c.rebalancer <- &Partition{num: e.TopicPartition.Partition, offset: e.TopicPartition.Offset}:
					}
					offsets[e.TopicPartition.Partition] = e.TopicPartition.Offset
				}

				select {
				case <-c.ctx.Done():
					return
				case c.processQueue <- e:
				}
			case kafka.Error:
				c.onError(c.ctx, e)
			}
		}
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
