package worker

import (
	"runtime"
)

// Source: http://marcio.io/2015/07/handling-1-million-requests-per-minute-with-golang/

// A Dispatcher of the workers list
type Dispatcher struct {
	state

	JobQueue chan ITask

	workerPool  chan chan ITask
	workersList []*worker
}

// NewDispatcher create new workers wispatcher
func NewDispatcher(countWorkers int) *Dispatcher {

	if countWorkers == 0 {
		countWorkers = runtime.NumCPU()
	}

	workerPool := make(chan chan ITask, countWorkers)
	workersList := make([]*worker, countWorkers, countWorkers)
	for i := 0; i < countWorkers; i++ {
		workersList[i] = newWorker(i, workerPool)
	}

	return &Dispatcher{
		workerPool:  workerPool,
		workersList: workersList,
		JobQueue:    make(chan ITask),
	}
}

// Run a tasks processor
func (d *Dispatcher) Run() {

	d.setIsClosed(false)

	go func() {
		defer func() {
			for i := 0; i < len(d.workersList); i++ {
				d.workersList[i].stop()
			}
		}()

		for i := 0; i < len(d.workersList); i++ {
			d.workersList[i].start()
		}

		for !d.getIsClosed() {
			jobChannel := <-d.workerPool
			jobChannel <- (<-d.JobQueue)
		}
	}()
}

// Close stops a tasks processor
func (d *Dispatcher) Close() error {
	d.setIsClosed(true)
	return nil
}
