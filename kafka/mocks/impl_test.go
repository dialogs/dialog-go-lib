package mocks

import (
	context "context"
	"fmt"
	"log"
	"testing"
	"time"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/require"
)

func TestMockClient(t *testing.T) {

	log.SetFlags(log.Lshortfile)

	const Topic = "test"

	w, r := NewClients(t, Topic)

	// test: message not found
	m, err := r.ReadMessage(context.Background())
	require.EqualError(t, err, "not found")
	require.Equal(t, kafkago.Message{}, m)

	// test: commit message in the empty order
	require.NoError(t, r.CommitMessages(context.Background(), kafkago.Message{
		Topic: Topic,
		Key:   []byte("key1"),
	}))

	// test: write message
	for i := 0; i < 2; i++ {
		require.NoError(t, w.WriteMessages(context.Background(), kafkago.Message{
			Key:   []byte(fmt.Sprintf("key%d", i)),
			Value: []byte(fmt.Sprintf("val%d", i)),
		}))
	}

	// test: read
	for i := 0; i < 2; i++ {
		m, err := r.ReadMessage(context.Background())
		require.NoError(t, err)
		require.WithinDuration(t, time.Now(), m.Time, time.Second)
		require.Equal(t,
			kafkago.Message{
				Topic:  Topic,
				Key:    []byte(fmt.Sprintf("key%d", i)),
				Value:  []byte(fmt.Sprintf("val%d", i)),
				Offset: int64(i),
				Time:   m.Time,
			},
			m)
	}

	require.NoError(t, r.CommitMessages(context.Background(), kafkago.Message{
		Topic: Topic,
		Key:   []byte("key0"),
	}))

	r.Close()

	m1, err := r.ReadMessage(context.Background())
	require.NoError(t, err)
	require.WithinDuration(t, time.Now(), m1.Time, time.Second)
	require.Equal(t,
		kafkago.Message{
			Topic:  Topic,
			Key:    []byte("key1"),
			Value:  []byte("val1"),
			Offset: 1,
			Time:   m1.Time,
		},
		m1)

	m2, err := r.FetchMessage(context.Background())
	require.EqualError(t, err, "not found")
	require.Equal(t, kafkago.Message{}, m2)
}
