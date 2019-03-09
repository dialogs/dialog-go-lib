package rand

import (
	"sort"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewValue(t *testing.T) {

	const (
		Threads    = 1000
		Iterations = 1000
		AllItems   = Threads * Iterations
	)

	res := make([]int64, AllItems)
	chUniq := make(chan int64, 1000)

	wgProcess := sync.WaitGroup{}

	wgProcess.Add(1)
	go func() {
		defer wgProcess.Done()

		var count int
		for val := range chUniq {
			res[count] = val
			count++
			if count == AllItems {
				close(chUniq)
			}
		}

		sort.Slice(res, func(i, j int) bool {
			return res[i] < res[j]
		})

		for i, val := range res {
			if i > 0 {
				require.NotEqual(t, res[i-1], val)
			}
		}
	}()

	for i := 0; i < Threads; i++ {
		wgProcess.Add(1)
		go func() {
			defer wgProcess.Done()

			for j := 0; j < Iterations; j++ {
				chUniq <- Int63()
			}
		}()
	}

	wgProcess.Wait()
}
