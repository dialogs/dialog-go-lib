package consumer

import (
	"container/heap"
	"context"
	"sync"

	"github.com/segmentio/kafka-go"
)

var nopCommitFunc = func(ctx context.Context, partition int, offset int64) {}

type GroupConfig struct {
	GroupID      string
	OnCommit     func(ctx context.Context, partition int, offset int64)
	OnError      func(err error)
	OnProcess    func(ctx context.Context, msg kafka.Message)
	QueueSize    int
	Reader       *kafka.Reader
	WorkersCount int
}

type Group struct {
	cancel       context.CancelFunc
	ctx          context.Context
	commitQueue  chan kafka.Message
	onCommit     func(ctx context.Context, partition int, offset int64)
	onError      func(err error)
	onProcess    func(ctx context.Context, msg kafka.Message)
	offsets      sync.Map
	processQueue chan kafka.Message
	reader       *kafka.Reader
	wg           sync.WaitGroup
	workersCount int
}

func NewGroupConfig() *GroupConfig {
	c := &GroupConfig{}

	c.QueueSize = 10
	c.WorkersCount = 10

	return c
}

func NewGroup(cfg *GroupConfig) *Group {
	if cfg.QueueSize == 0 {
		panic("consumer group queue size is 0")
	}

	if cfg.WorkersCount == 0 {
		panic("consumer group queue size is 0")
	}

	if cfg.OnError == nil {
		panic("on error callback is nil")
	}

	if cfg.OnProcess == nil {
		panic("on process callback is nil")
	}

	if cfg.Reader == nil {
		panic("reader is nil")
	}

	if len(cfg.Reader.Config().GroupID) == 0 {
		panic("reader group id is empty")
	}

	var onCommit func(ctx context.Context, partition int, offset int64)
	if cfg.OnCommit != nil {
		onCommit = cfg.OnCommit
	} else {
		onCommit = nopCommitFunc
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Group{
		cancel:       cancel,
		ctx:          ctx,
		commitQueue:  make(chan kafka.Message, cfg.QueueSize),
		onCommit:     onCommit,
		onProcess:    cfg.OnProcess,
		processQueue: make(chan kafka.Message, cfg.QueueSize),
		reader:       cfg.Reader,
		workersCount: cfg.WorkersCount,
	}
}

func (g *Group) Start() {
	g.wg.Add(1)
	go func() {
		g.readLoop()
		g.wg.Done()
	}()

	g.wg.Add(1)
	go func() {
		g.commitLoop()
		g.wg.Done()
	}()

	g.wg.Add(g.workersCount)
	for i := 0; i < g.workersCount; i++ {
		go func() {
			g.processLoop()
			g.wg.Done()
		}()
	}

}

func (g *Group) Stop() {
	g.cancel()
	g.wg.Wait()
}

func (g *Group) getOffset(partition int) int64 {
	val, _ := g.offsets.Load(partition)

	return val.(int64)
}

func (g *Group) commitLoop() {
	targets := make([]kafka.Message, 0)
	partitions := make(map[int]*MessageHeap)
	offsets := make(map[int]int64)

	for {
		select {
		case <-g.ctx.Done():
			return
		case msg := <-g.commitQueue:
			expectedOffset, ok := offsets[msg.Partition]
			if !ok {
				expectedOffset = g.getOffset(msg.Partition)
				offsets[msg.Partition] = expectedOffset
			}

			h, ok := partitions[msg.Partition]
			if !ok {
				h = new(MessageHeap)
				partitions[msg.Partition] = h
			}

			heap.Push(h, msg)

			// extract longest increasing subsequence starting from expected offset
			for {
				if h.Len() != 0 && (*h)[0].Offset == expectedOffset {
					expectedOffset++
					offsets[msg.Partition] = expectedOffset
					targets = append(targets, heap.Pop(h).(kafka.Message))
				} else {
					break
				}
			}
			if len(targets) > 0 {
				if err := g.reader.CommitMessages(g.ctx, targets...); err != nil {
					if err == context.Canceled {
						return
					}
					g.onError(err)
					continue
				}

				g.onCommit(g.ctx, msg.Partition, expectedOffset)

				targets = targets[:0]
			}
		}
	}
}

func (g *Group) commitMessage(m kafka.Message) {
	select {
	case <-g.ctx.Done():
		return
	case g.commitQueue <- m:

	}
}

func (g *Group) processLoop() {
	for {
		select {
		case <-g.ctx.Done():
			return
		case msg := <-g.processQueue:
			g.onProcess(g.ctx, msg)
			g.commitMessage(msg)
		}
	}
}

func (g *Group) readLoop() {
	for {
		msg, ok := g.readNext()
		if !ok {
			return
		}

		select {
		case <-g.ctx.Done():
			return
		case g.processQueue <- msg:
		}
	}
}

func (g *Group) readNext() (kafka.Message, bool) {
	msg, err := g.reader.FetchMessage(g.ctx)
	if err != nil {
		if err == context.Canceled {
			return kafka.Message{}, false
		}

		g.onError(err)
	}
	return msg, true
}

func (g *Group) setOffset(partition int, offset int64) {
	g.offsets.LoadOrStore(partition, offset)
}
