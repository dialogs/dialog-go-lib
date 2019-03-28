package worker

import "sync/atomic"

type state struct {
	isClosed int32
}

func (s *state) getIsClosed() bool {
	return atomic.LoadInt32(&s.isClosed) > 0
}

func (s *state) setIsClosed(isClosed bool) {

	var val int32
	if isClosed {
		val = 1
	}

	atomic.StoreInt32(&s.isClosed, val)
}
