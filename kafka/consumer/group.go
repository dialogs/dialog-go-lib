package consumer

import (
	"container/list"
	"context"
	"runtime"
	"sync"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type GroupConfig struct {
	Config  *Config `mapstructure:"config"`
	Workers int     `mapstructure:"workers"`
}

type Group struct {
	consumerConfig *Config
	workers        int
	logger         *zap.Logger
	wg             sync.WaitGroup
	ctx            context.Context
	ctxCancel      func()
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

	ctxDone, ctxDoneCancel := context.WithCancel(context.Background())

	return &Group{
		consumerConfig: cfg.Config,
		workers:        workers,
		logger:         logger,
		ctx:            ctxDone,
		ctxCancel:      ctxDoneCancel,
	}, nil
}

func (g *Group) Start() (err error) {

	g.logger.Info("wait for start")
	g.wg.Wait()

	select {
	case <-g.ctx.Done():
		return errors.Wrap(err, "consumers group already closed")
	default:
		// ok
	}

	consumerList := list.New()
	for i := 0; i < g.workers; i++ {
		c, err := New(g.consumerConfig, g.logger)
		if err != nil {
			return errors.Wrap(err, "failed to start consumer group")
		}

		consumerList.PushBack(c)
	}

	defer func() {
		for item := consumerList.Front(); item != nil; item = item.Next() {
			item.Value.(*Consumer).Stop()
		}
	}()

	retval := make(chan error, consumerList.Len()+1) // +1 context closed

	go func() {
		<-g.ctx.Done()
		retval <- nil
	}()

	for item := consumerList.Front(); item != nil; item = item.Next() {
		g.wg.Add(1)
		go func(worker *Consumer) {
			defer g.wg.Done()
			retval <- worker.Start()

		}(item.Value.(*Consumer))
	}

	return <-retval
}

func (g *Group) Stop() {
	g.ctxCancel()
	g.wg.Wait()
}
