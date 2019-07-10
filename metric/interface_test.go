package metric

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/dialogs/dialog-go-lib/metric/mock"
	"github.com/stretchr/testify/require"
)

func TestInterface(t *testing.T) {

	// test: mock object
	require.NotNil(t, (ICounter)(mock.NewCounter()))
	// test: prometheus object
	require.NotNil(t,
		(ICounter)(prometheus.NewCounter(prometheus.CounterOpts{})))

	// test: mock object
	require.NotNil(t, (IObserver)(mock.NewObserver()))
	// test: prometheus object
	require.NotNil(t,
		(IObserver)(prometheus.NewHistogram(prometheus.HistogramOpts{})))
	require.NotNil(t,
		(IObserver)(
			prometheus.NewHistogramVec(prometheus.HistogramOpts{}, []string{}).
				With(prometheus.Labels{})))

}
