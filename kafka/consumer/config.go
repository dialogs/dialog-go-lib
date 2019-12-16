package consumer

import (
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
)

type Config struct {
	ConfigMap            *kafka.ConfigMap
	CommitOffsetCount    int
	CommitOffsetDuration time.Duration
	OnCommit             FuncOnCommit
	OnError              FuncOnError
	OnProcess            FuncOnProcess
	OnRevoke             FuncOnRevoke
	OnRebalance          FuncOnRebalance
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
