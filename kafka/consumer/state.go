package consumer

import "sync"

type StateEvent int

const (
	StateRun StateEvent = iota
	StateClosed
)

type IStateObserver interface {
	HandleStateEvent(StateEvent)
}

type observable struct {
	observers []IStateObserver
	mu        sync.RWMutex
}

func newObservable() *observable {
	return &observable{
		observers: make([]IStateObserver, 0),
	}
}

func (o *observable) AddStateObserver(v ...IStateObserver) {
	o.mu.Lock()
	o.observers = append(o.observers, v...)
	o.mu.Unlock()
}

func (o *observable) RemoveStateObserver(v IStateObserver) {
	o.mu.Lock()
	defer o.mu.Unlock()

	for i := range o.observers {
		if o.observers[i] == v {
			lastIndex := len(o.observers) - 1
			if i < lastIndex {
				copy(o.observers[i:], o.observers[i+1:])
			}

			o.observers = o.observers[:lastIndex]
			return
		}
	}
}

func (o *observable) notify(e StateEvent) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	for _, item := range o.observers {
		item.HandleStateEvent(e)
	}
}
