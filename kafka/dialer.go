package kafka

import (
	kafkago "github.com/segmentio/kafka-go"
)

func newDialer(config *Config) *kafkago.Dialer {
	return &kafkago.Dialer{
		DualStack: config.DualStack,
		Timeout:   config.Timeout,
		TLS:       config.TLSConfig,
	}
}
