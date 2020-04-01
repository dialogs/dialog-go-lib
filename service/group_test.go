package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGroupWithCancel(t *testing.T) {

	const TimeoutStep = time.Millisecond * 10

	fnWithError := func(ctx context.Context) error {
		time.Sleep(TimeoutStep)
		return errors.New("task 0")
	}

	tasks := []GroupTask{fnWithError}
	for i := 0; i < 100; i++ {
		fn := func(num int) GroupTask {
			return func(ctx context.Context) error {
				<-ctx.Done()
				time.Sleep(TimeoutStep * time.Duration(num))
				return fmt.Errorf("task %d", num)
			}
		}(len(tasks))

		tasks = append(tasks, fn)
	}

	var (
		count int
		once  sync.Once
	)

	chErr, cancel := RunGroup(tasks...)
	for err := range chErr {
		once.Do(cancel)

		require.EqualError(t, err, fmt.Sprintf("task %d", count))
		count++
	}

	require.Equal(t, len(tasks), count)
}

func TestGroupWaitAll(t *testing.T) {

	const TimeoutStep = time.Millisecond * 10

	tasks := make([]GroupTask, 0)
	for i := 0; i < 100; i++ {
		fn := func(num int) GroupTask {
			return func(_ context.Context) error {
				time.Sleep(TimeoutStep * time.Duration(num))
				return fmt.Errorf("task %d", num)
			}
		}(len(tasks))

		tasks = append(tasks, fn)
	}

	var count int
	chErr, _ := RunGroup(tasks...)
	for err := range chErr {
		require.EqualError(t, err, fmt.Sprintf("task %d", count))
		count++
	}

	require.Equal(t, len(tasks), count)
}
