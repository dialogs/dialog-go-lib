package kafka

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNormalizeConfig(t *testing.T) {
	m := kafka.ConfigMap{}
	require.NoError(t, m.SetKey("-strange-key", "1"))
	require.NoError(t, m.SetKey("stranger-key-", "2"))
	require.NoError(t, m.SetKey("totally-legit-key", "3"))

	res := NormalizeConfig(m)
	sk1, err := res.Get("strange.key", "NO VALUE")
	require.NoError(t, err)
	require.Equal(t, "1", sk1)

	ssk1, err := res.Get(".strange.key", "NO VALUE")
	require.NoError(t, err)
	require.Equal(t, "NO VALUE", ssk1)

	sk2, err := res.Get("stranger.key", "NO VALUE")
	require.NoError(t, err)
	require.Equal(t, "2", sk2)

	ssk2, err := res.Get("stranger.key.", "NO VALUE")
	require.NoError(t, err)
	require.Equal(t, "NO VALUE", ssk2)

	sk3, err := res.Get("totally.legit.key", "NO VALUE")
	require.NoError(t, err)
	require.Equal(t, "3", sk3)
}
