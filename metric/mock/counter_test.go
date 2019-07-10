package mock

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCounter(t *testing.T) {

	c := NewCounter()

	c.Inc()
	require.Equal(t, uint64(1), c.Get())

	c.Inc()
	require.Equal(t, uint64(2), c.Get())

	func() {
		defer func() {
			e := recover()
			require.Equal(t, "invalid value: -0.1", e)
		}()

		c.Add(-0.1)
	}()

	c.Add(2)
	require.Equal(t, uint64(4), c.Get())
}
