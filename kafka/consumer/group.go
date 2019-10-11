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
	consumers *list.List
	logger    *zap.Logger
	ctx       context.Context
	ctxCancel func()
	mu        sync.RWMutex
	wg        sync.WaitGroup
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

	g.mu.Lock()
	defer g.mu.Unlock()

	select {
	case <-g.ctx.Done():
		return errors.New("consumers group already closed")
	default:
		// ok
	}

	g.logger.Info("starting ...")
	retval := make(chan error, g.consumers.Len()+1) // +1 context closed

	defer g.ctxCancel()

	g.wg.Add(g.consumers.Len())
	for item := g.consumers.Front(); item != nil; item = item.Next() {
		go func(c *Consumer) {
			defer g.wg.Done()

			retval <- newGroupItem(c, g.ctx).Start()
		}(item.Value.(*Consumer))
	}

	go func() {
		<-g.ctx.Done()
		retval <- nil // success shutdown
	}()

	g.logger.Info("success start")

	return <-retval
}

func (g *Group) Stop() {

	g.ctxCancel()

	g.mu.Lock() // protection for WaitGroup data race
	defer g.mu.Unlock()

	g.wg.Wait()
}
