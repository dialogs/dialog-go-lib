package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
	kafkago "github.com/segmentio/kafka-go"
)

// IReader interface of kafka reader client
type IReader interface {
	ReadMessage(ctx context.Context) (kafkago.Message, error)
	FetchMessage(ctx context.Context) (kafkago.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafkago.Message) error
	Close() (err error)
}

// NewReader create original kafka reader client
func NewReader(groupID, topic string, config *Config) *kafkago.Reader {

	dialer := newDialer(config)
	rConf := kafkago.ReaderConfig{
		Brokers: config.Brokers,
		Dialer:  dialer,
		GroupID: groupID,
		Topic:   topic,
	}

	return kafka.NewReader(rConf)
}
