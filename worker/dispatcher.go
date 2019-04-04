package worker

import (
	"context"
	"runtime"
)

// Source: http://marcio.io/2015/07/handling-1-million-requests-per-minute-with-golang/

// A Dispatcher of the workers list
type Dispatcher struct {
	jobQueue     chan ITask
	countWorkers int
}

// NewDispatcher create new workers wispatcher
func NewDispatcher(countWorkers int) *Dispatcher {

	if countWorkers == 0 {
		countWorkers = runtime.NumCPU()
	}

	return &Dispatcher{
		countWorkers: countWorkers,
		jobQueue:     make(chan ITask),
	}
}

// Invoke adds task to the queue
func (d *Dispatcher) Invoke(task ITask) {
	d.jobQueue <- task
}

// Run a tasks processor
func (d *Dispatcher) Run(ctx context.Context) {

	workerPool := make(chan struct{}, d.countWorkers)
	for i := 0; i < cap(workerPool); i++ {
		workerPool <- struct{}{}
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case <-workerPool:
				select {
				case <-ctx.Done():
					return

				case task := <-d.jobQueue:
					go func() {
						defer func() {
							workerPool <- struct{}{}
						}()

						task.Invoke()
					}()
				}
			}
		}
	}()
}
