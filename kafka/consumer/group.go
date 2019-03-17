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
	offsets      sync.Map
	processQueue chan kafka.Message
	reader       *kafka.Reader
	workersCount int
	wg           sync.WaitGroup
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

func (c *Consumer) getOffset(partition int) int64 {
	//always expect not nil val
	val, _ := c.offsets.Load(partition)

	return val.(int64)
}

func (c *Consumer) setOffset(partition int, offset int64) {
	c.offsets.LoadOrStore(partition, offset)
}

func (c *Consumer) commitLoop() {
	targets := make([]kafka.Message, 0)
	partitions := make(map[int]*MessageHeap)
	offsets := make(map[int]int64)

	for {
		select {
		case <-c.ctx.Done():
			return
		case msg := <-c.commitQueue:
			expectedOffset, ok := offsets[msg.Partition]
			if !ok {
				expectedOffset = c.getOffset(msg.Partition)
				offsets[msg.Partition] = expectedOffset
			}

			h, ok := partitions[msg.Partition]
			if !ok {
				h = new(MessageHeap)
				partitions[msg.Partition] = h
			}

			heap.Push(h, msg)

			// extract longest increasing subsequence starting from expected offset
			for h.Len() != 0 && (*h)[0].Offset == expectedOffset {
				expectedOffset++
				offsets[msg.Partition] = expectedOffset
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

				c.onCommit(c.ctx, msg.Partition, expectedOffset, len(targets))

				targets = targets[:0]
			}
		}
	}
}

func (c *Consumer) commitMessage(m kafka.Message) {
	select {
	case <-c.ctx.Done():
		return
	case c.commitQueue <- m:

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
	for {
		msg, ok := c.readNext()
		if !ok {
			return
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

	c.setOffset(msg.Partition, msg.Offset)

	return msg, true
}
