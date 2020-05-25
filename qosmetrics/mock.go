package qosmetrics

import (
	"time"

	"github.com/dialogs/dialog-go-lib/metric/mock"
)

type mockMetric struct {
	Observer        *mock.Observer
	ErrorCounter    *mock.Counter
	ConsumedCounter *mock.Counter
}

func NewMockMetric() *mockMetric {
	return &mockMetric{
		Observer:        mock.NewObserver(),
		ErrorCounter:    mock.NewCounter(),
		ConsumedCounter: mock.NewCounter(),
	}
}

func (s *mockMetric) IncConsumed(methodName string) {
	s.ConsumedCounter.Inc()
}

func (s *mockMetric) IncErrored(methodName string) {
	s.ErrorCounter.Inc()
}

func (s *mockMetric) ObserveLatency(methodName string, start time.Time) {
	s.Observer.Observe(time.Since(start).Seconds())
}
