package service

import (
	"context"
	"sync"
)

type GroupTask func(ctx context.Context) error

func RunGroup(tasks ...GroupTask) (_ <-chan error, cancel func()) {

	var (
		wg    sync.WaitGroup
		ctx   context.Context
		chErr = make(chan error)
	)

	ctx, cancel = context.WithCancel(context.Background())

	for _, task := range tasks {
		wg.Add(1)
		go func(fn GroupTask) {
			defer wg.Done()
			chErr <- fn(ctx)
		}(task)
	}

	go func() {
		wg.Wait()
		close(chErr)
	}()

	return chErr, cancel
}
