package worker

import (
	"context"
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

type TaskResult struct {
	Thread int
	Number int
}

type Task struct {
	Thread  int
	Number  int
	Results chan *TaskResult
}

func (t *Task) Invoke() {
	t.Results <- &TaskResult{
		Thread: t.Thread,
		Number: t.Number,
	}
}

func TestWorkers(t *testing.T) {

	const (
		CountThreads = 1000
		CountTasks   = 1000
	)

	results := make(chan *TaskResult, CountTasks)

	ctx, cancel := context.WithCancel(context.Background())

	dispatcher := NewDispatcher(0)
	dispatcher.Run(ctx)

	require.Equal(t, runtime.NumCPU(), dispatcher.countWorkers)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		const resultValue = CountTasks * CountThreads
		var totalCount int

		for val := range results {
			require.NotEqual(t, 0, val.Number)
			require.NotEqual(t, 0, val.Thread)

			totalCount++
			if totalCount == resultValue {
				return
			}
		}
	}()

	for threadIdx := 1; threadIdx <= CountThreads; threadIdx++ {
		go func(thread int) {
			for taskIdx := 1; taskIdx <= CountTasks; taskIdx++ {
				task := &Task{
					Thread:  thread,
					Number:  taskIdx,
					Results: results,
				}

				dispatcher.Invoke(task)
			}
		}(threadIdx)
	}

	wg.Wait()

	results <- nil
	require.Nil(t, <-results)

	cancel()

}
