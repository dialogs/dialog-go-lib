package consumer

import (
	"context"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
)

type Config struct {
	ConfigMap            *kafka.ConfigMap
	CommitOffsetCount    int
	CommitOffsetDuration time.Duration
	OnCommit             func(ctx context.Context, partition int32, offset kafka.Offset, committed int)
	OnError              func(ctx context.Context, err error)
	OnProcess            func(ctx context.Context, msg *kafka.Message)
	Topics               []string
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Check() error {

	if c.OnError == nil {
		return errors.New("on error callback is nil")
	}

	if c.OnProcess == nil {
		return errors.New("on process callback is nil")
	}

	if len(c.Topics) == 0 {
		return errors.New("topics is empty")
	}

	if c.ConfigMap == nil {
		return errors.New("reader config is nil")
	}

	return nil
}
