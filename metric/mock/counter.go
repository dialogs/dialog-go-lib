package mock

import (
	"fmt"
	"sync/atomic"
)

type Counter struct {
	val uint64
}

func NewCounter() *Counter {
	return &Counter{}
}

func (c *Counter) Inc() {
	atomic.AddUint64(&c.val, 1)
}

func (c *Counter) Add(val float64) {

	if val < 0 {
		panic(fmt.Sprintf("invalid value: %v", val))
	}

	atomic.AddUint64(&c.val, uint64(val))
}

func (c *Counter) Get() uint64 {
	return atomic.LoadUint64(&c.val)
}
