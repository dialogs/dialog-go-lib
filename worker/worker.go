package worker

type worker struct {
	state

	id         int
	workerPool chan chan ITask
	jobChannel chan ITask
}

func newWorker(id int, workerPool chan chan ITask) *worker {
	return &worker{
		id:         id,
		workerPool: workerPool,
		jobChannel: make(chan ITask),
	}
}

func (w *worker) start() {
	w.setIsClosed(false)

	go func() {
		for !w.getIsClosed() {
			w.workerPool <- w.jobChannel
			job := <-w.jobChannel
			job.Invoke()
		}
	}()
}

func (w *worker) stop() {
	w.setIsClosed(true)
}
