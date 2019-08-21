package consumer

import (
	"context"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/require"
)

func TestConfigCheck(t *testing.T) {

	require.EqualError(t,
		(&Config{}).Check(),
		"on error callback is nil")

	require.EqualError(t,
		(&Config{
			OnError: func(context.Context, error) {},
		}).Check(),
		"on process callback is nil")

	require.EqualError(t,
		(&Config{
			OnError:   func(context.Context, error) {},
			OnProcess: func(context.Context, *kafka.Message) {},
		}).Check(),
		"topics is empty")

	require.EqualError(t,
		(&Config{
			OnError:   func(context.Context, error) {},
			OnProcess: func(context.Context, *kafka.Message) {},
			Topics:    []string{"a"},
		}).Check(),
		"reader config is nil")

	require.NoError(t,
		(&Config{
			OnError:   func(context.Context, error) {},
			OnProcess: func(context.Context, *kafka.Message) {},
			Topics:    []string{"a"},
			ConfigMap: &kafka.ConfigMap{"bootstrap.servers": "b1,b2,b3"},
		}).Check())

}
