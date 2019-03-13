package mocks

import (
	context "context"
	"errors"
	"testing"

	kafkago "github.com/segmentio/kafka-go"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type WriterMock struct {
	IWriter

	topic string
	buf   *buffer
}

type ReaderMock struct {
	IReader

	topic string
	buf   *buffer
}

func NewClients(t *testing.T, topic string) (*WriterMock, *ReaderMock) {
	buf := newBuffer()
	return newWriter(t, topic, buf), newReader(t, topic, buf)
}

func newReader(t *testing.T, topic string, buf *buffer) *ReaderMock {

	client := &ReaderMock{
		topic: topic,
		buf:   buf,
	}

	getter := func() (
		func(ctx context.Context) kafkago.Message,
		func(ctx context.Context) error,
	) {
		var val *kafkago.Message

		return func(ctx context.Context) kafkago.Message {
				val = client.buf.Get(client.topic)
				if val != nil {
					return *val
				}

				return kafkago.Message{}
			}, func(ctx context.Context) error {
				if val != nil {
					return nil
				}

				return errors.New("not found")
			}
	}

	client.On("ReadMessage",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
	).Return(
		getter(),
	)

	client.On("FetchMessage",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
	).Return(
		getter(),
	)

	client.On("CommitMessages",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(msgs kafkago.Message) bool { return true }),
	).Return(
		func(ctx context.Context, msgs ...kafkago.Message) error {
			for _, m := range msgs {
				if m.Topic == "" {
					return errors.New("Received unexpected error: unable to commit offsets for group, : [3] Unknown Topic Or Partition: the request is for a topic or partition that does not exist on this broker")
				}
				client.buf.Commit(client.topic, m)
			}

			return nil
		},
	)

	client.On("Close").Return(
		func() error {
			client.buf.Close()
			return nil
		},
	)

	return client
}

func newWriter(t *testing.T, topic string, buf *buffer) *WriterMock {

	client := &WriterMock{
		topic: topic,
		buf:   buf,
	}

	client.On("WriteMessages",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(kafkago.Message) bool { return true }),
	).Return(
		func(ctx context.Context, msgs ...kafkago.Message) error {
			require.NotEmpty(t, msgs)
			for i := range msgs {
				m := &msgs[i]
				if m.Topic == "" {
					m.Topic = client.topic
				}

				require.Equal(t, client.topic, m.Topic)
				require.NotEmpty(t, m.Key)
				require.NotEmpty(t, m.Value)

				client.buf.Add(client.topic, *m)
			}

			return nil
		},
	)

	client.On("Close").Return(
		func() error {
			return nil
		},
	)

	return client
}
