package consumer

import (
	"context"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestConfigNew(t *testing.T) {
	require.Equal(t, &Config{}, NewConfig())
}

func TestConfigCheck(t *testing.T) {

	require.EqualError(t,
		(&Config{}).Check(),
		"on error callback is nil")

	require.EqualError(t,
		(&Config{
			OnError: func(context.Context, *zap.Logger, error) {},
		}).Check(),
		"on process callback is nil")

	require.EqualError(t,
		(&Config{
			OnError:   func(context.Context, *zap.Logger, error) {},
			OnProcess: func(context.Context, *zap.Logger, *kafka.Message, DelayI) error { return nil },
		}).Check(),
		"topics is empty")

	require.EqualError(t,
		(&Config{
			OnError:   func(context.Context, *zap.Logger, error) {},
			OnProcess: func(context.Context, *zap.Logger, *kafka.Message, DelayI) error { return nil },
			Topics:    []string{"a"},
		}).Check(),
		"reader config is nil")

	require.NoError(t,
		(&Config{
			OnError:   func(context.Context, *zap.Logger, error) {},
			OnProcess: func(context.Context, *zap.Logger, *kafka.Message, DelayI) error { return nil },
			Topics:    []string{"a"},
			ConfigMap: &kafka.ConfigMap{"bootstrap.servers": "b1,b2,b3"},
		}).Check())

}
