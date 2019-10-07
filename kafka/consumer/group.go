package consumer

import (
	"container/list"
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type GroupConfig struct {
	Config  *Config `mapstructure:"config"`
	Workers int     `mapstructure:"workers"`
}

type Group struct {
	consumers *list.List
	logger    *zap.Logger
	wg        sync.WaitGroup
	ctx       context.Context
	ctxCancel func()
	started   int32 // for graceful shutdown
}

func NewGroup(cfg GroupConfig, logger *zap.Logger) (*Group, error) {

	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	logger = logger.With(zap.String("consumers group", id.String()))

	workers := cfg.Workers
	if workers <= 0 {
		workers = runtime.NumCPU()
	}

	consumers := list.New()
	for i := 0; i < workers; i++ {
		c, err := New(cfg.Config, logger)
		if err != nil {
			return nil, errors.Wrap(err, "failed to start consumer group")
		}

		consumers.PushBack(c)
	}

	ctxDone, ctxDoneCancel := context.WithCancel(context.Background())

	return &Group{
		consumers: consumers,
		logger:    logger,
		ctx:       ctxDone,
		ctxCancel: ctxDoneCancel,
	}, nil
}

func (g *Group) Start() error {

	g.logger.Info("wait for start")
	g.wg.Wait()

	select {
	case <-g.ctx.Done():
		return errors.New("consumers group already closed")
	default:
		// ok
	}

	g.logger.Info("starting ...")

	atomic.StoreInt32(&g.started, int32(g.consumers.Len()))
	defer func() {
		g.ctxCancel()

		for item := g.consumers.Front(); item != nil; item = item.Next() {
			item.Value.(*Consumer).Stop()
			atomic.AddInt32(&g.started, -1)
		}
	}()

	retval := make(chan error, g.consumers.Len()+1) // +1 context closed

	go func() {
		<-g.ctx.Done()
		retval <- nil
	}()

	for item := g.consumers.Front(); item != nil; item = item.Next() {
		g.wg.Add(1)
		go func(worker *Consumer) {
			defer g.wg.Done()
			retval <- worker.Start()

		}(item.Value.(*Consumer))
	}

	g.logger.Info("success start")

	return <-retval
}

func (g *Group) Stop() {
	g.ctxCancel()

	{
		// graceful shutdown
		tc := time.NewTicker(time.Millisecond)
		defer tc.Stop()
		for range tc.C {
			if atomic.LoadInt32(&g.started) == 0 {
				break
			}
		}
	}

	g.wg.Wait()
}
