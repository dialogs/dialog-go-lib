package mock

import (
	"container/list"
	"sync"
)

type Observer struct {
	list *list.List
	mu   sync.RWMutex
}

func NewObserver() *Observer {
	return &Observer{
		list: list.New(),
	}
}

func (o *Observer) Observe(val float64) {

	o.mu.Lock()
	o.list.PushBack(val)
	o.mu.Unlock()
}

func (o *Observer) GetSlice() []float64 {

	o.mu.RLock()
	defer o.mu.RUnlock()

	retval := make([]float64, 0, o.list.Len())
	for e := o.list.Front(); e != nil; e = e.Next() {
		retval = append(retval, e.Value.(float64))
	}

	return retval
}

func (o *Observer) GetAvg() float64 {

	o.mu.RLock()
	defer o.mu.RUnlock()

	if o.list.Len() == 0 {
		return 0
	}

	var sum float64
	for e := o.list.Front(); e != nil; e = e.Next() {
		sum += e.Value.(float64)
	}

	return sum / float64(o.list.Len())
}
