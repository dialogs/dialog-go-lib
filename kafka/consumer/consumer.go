package consumer

import (
	"container/heap"
	"context"
	"errors"
	"sync"

	"github.com/segmentio/kafka-go"
)

var nopCommitFunc = func(ctx context.Context, partition int, offset int64, committed int) {}

type Config struct {
	OnCommit     func(ctx context.Context, partition int, offset int64, committed int)
	OnError      func(err error)
	OnProcess    func(ctx context.Context, msg kafka.Message)
	QueueSize    int
	Reader       *kafka.Reader
	WorkersCount int
}

type Consumer struct {
	cancel       context.CancelFunc
	ctx          context.Context
	commitQueue  chan kafka.Message
	onCommit     func(ctx context.Context, partition int, offset int64, committed int)
	onError      func(err error)
	onProcess    func(ctx context.Context, msg kafka.Message)
	processQueue chan kafka.Message
	reader       *kafka.Reader
	rebalancer   chan *Partition
	workersCount int
	wg           sync.WaitGroup
}

type Partition struct {
	number int
	offset int64
}

func NewConfig() *Config {
	return &Config{
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

	if cfg.Reader == nil {
		return nil, errors.New("reader is nil")
	}

	var onCommit func(ctx context.Context, partition int, offset int64, committed int)
	if cfg.OnCommit != nil {
		onCommit = cfg.OnCommit
	} else {
		onCommit = nopCommitFunc
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Consumer{
		cancel:       cancel,
		ctx:          ctx,
		commitQueue:  make(chan kafka.Message, cfg.QueueSize),
		onCommit:     onCommit,
		onProcess:    cfg.OnProcess,
		processQueue: make(chan kafka.Message, cfg.QueueSize),
		reader:       cfg.Reader,
		rebalancer:   make(chan *Partition),
		workersCount: cfg.WorkersCount,
	}, nil
}

func (c *Consumer) Start() {
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

}

func (c *Consumer) Stop() {
	c.cancel()
	c.wg.Wait()
	c.reader.Close()
}

func (c *Consumer) commitLoop() {
	targets := make([]kafka.Message, 0)
	partitions := make(map[int]*MessageHeap)
	offsets := make(map[int]int64)

	for {
		select {
		case <-c.ctx.Done():
			return
		case partition := <-c.rebalancer:
			expectedOffset, ok := offsets[partition.number]
			if ok {
				h := partitions[partition.number]
				for h.Len() != 0 && (*h)[0].Offset == expectedOffset {
					expectedOffset++
					heap.Pop(h)
				}
			}

			offsets[partition.number] = partition.offset

		case msg := <-c.commitQueue:
			expectedOffset := offsets[msg.Partition]

			h, ok := partitions[msg.Partition]
			if !ok {
				h = new(MessageHeap)
				partitions[msg.Partition] = h
			}

			heap.Push(h, msg)

			// extract longest increasing subsequence starting from expected offset
			for h.Len() != 0 && (*h)[0].Offset == expectedOffset {
				expectedOffset++
				targets = append(targets, heap.Pop(h).(kafka.Message))
			}
			if len(targets) > 0 {
				if err := c.reader.CommitMessages(c.ctx, targets...); err != nil {
					if err == context.Canceled {
						return
					}
					c.onError(err)
					continue
				}

				offsets[msg.Partition] = expectedOffset

				c.onCommit(c.ctx, msg.Partition, expectedOffset, len(targets))

				targets = targets[:0]
			}
		}
	}
}

func (c *Consumer) commitMessage(msg kafka.Message) {
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
	offsets := make(map[int]int64)

	for {
		msg, ok := c.readNext()
		if !ok {
			return
		}

		offset, ok := offsets[msg.Partition]
		if !ok || offset+1 < msg.Offset {
			offsets[msg.Partition] = msg.Offset

			select {
			case <-c.ctx.Done():
				return
			case c.rebalancer <- &Partition{number: msg.Partition, offset: msg.Offset}:

			}
		}

		select {
		case <-c.ctx.Done():
			return
		case c.processQueue <- msg:
		}
	}
}

func (c *Consumer) readNext() (kafka.Message, bool) {
	msg, err := c.reader.FetchMessage(c.ctx)
	if err != nil {
		if err == context.Canceled {
			return kafka.Message{}, false
		}

		c.onError(err)
	}

	return msg, true
}
