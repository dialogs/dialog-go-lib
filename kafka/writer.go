package kafka

import (
	"context"

	kafkago "github.com/segmentio/kafka-go"
)

// IWriter interface of kafka writer client
type IWriter interface {
	WriteMessages(ctx context.Context, msgs ...kafkago.Message) error
	Close() (err error)
}

// NewWriter create original kafka writer client
func NewWriter(topic string, config *Config) *kafkago.Writer {

	dialer := newDialer(config)
	wConf := kafkago.WriterConfig{
		Brokers: config.Brokers,
		Dialer:  dialer,
		Topic:   topic,
	}

	return kafkago.NewWriter(wConf)
}
