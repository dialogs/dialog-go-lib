package mock

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestObserver(t *testing.T) {

	o := NewObserver()

	o.Observe(1)
	require.Equal(t, []float64{1}, o.GetSlice())
	require.Equal(t, float64(1), o.GetAvg())

	o.Observe(1)
	require.Equal(t, []float64{1, 1}, o.GetSlice())
	require.Equal(t, float64(1), o.GetAvg())

	o.Observe(4)
	require.Equal(t, []float64{1, 1, 4}, o.GetSlice())
	require.Equal(t, float64(2), o.GetAvg())
}
