package consumer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type TestObserver struct {
	Event StateEvent
}

func (t *TestObserver) HandleStateEvent(e StateEvent) {
	t.Event = e
}

func TestObserverAddRemove(t *testing.T) {

	target1 := &TestObserver{}
	target2 := &TestObserver{}
	target3 := &TestObserver{}
	target4 := &TestObserver{}

	o := newObservable()
	o.AddStateObserver(target1, target2, target3, target4)

	require.Equal(t,
		&observable{
			observers: []IStateObserver{target1, target2, target3, target4},
		},
		o)

	// test: remove from the middle
	o.RemoveStateObserver(target2)
	require.Equal(t,
		&observable{
			observers: []IStateObserver{target1, target3, target4},
		},
		o)

	// test: remove first
	o.RemoveStateObserver(target1)
	require.Equal(t,
		&observable{
			observers: []IStateObserver{target3, target4},
		},
		o)

	// test: remove last
	o.RemoveStateObserver(target4)
	require.Equal(t,
		&observable{
			observers: []IStateObserver{target3},
		},
		o)

	o.RemoveStateObserver(target3)
	require.Equal(t,
		&observable{
			observers: []IStateObserver{},
		},
		o)
}

func TestObserverNotify(t *testing.T) {

	o := newObservable()

	const Count = 10
	targets := make([]*TestObserver, 0, Count)
	for i := 0; i < Count; i++ {
		item := &TestObserver{}
		targets = append(targets, item)
		o.AddStateObserver(item)
	}

	for _, e := range []StateEvent{StateRun, StateClosed} {
		o.notify(e)
		for i := 0; i < Count; i++ {
			require.Equal(t, e, targets[i].Event)
		}
	}
}
