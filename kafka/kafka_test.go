package kafka

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	kafkago "github.com/segmentio/kafka-go"
)

func TestKafka(t *testing.T) {

	conf := &Config{
		Brokers: []string{
			"localhost:9092",
		},
		GroupID:   "kafkatest",
		Timeout:   time.Second,
		DualStack: true,
	}

	const Topic = "TestKafka1"

	conn, err := (&kafkago.Dialer{
		Resolver: &net.Resolver{},
	}).DialLeader(context.Background(), "tcp", conf.Brokers[0], Topic, 0)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, conn.DeleteTopics(Topic))
	}()

	require.NoError(t, conn.CreateTopics(kafkago.TopicConfig{
		Topic:             Topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}))

	var (
		r IReader = NewReader(Topic, conf)
		w IWriter = NewWriter(Topic, conf)
	)

	// Check mock:
	// w, r = mocks.NewClients(t, Topic)

	defer func() {
		require.NoError(t, r.Close())
		require.NoError(t, w.Close())
	}()

	for i := 0; i < 1; i++ {
		func() {
			text := fmt.Sprintf("message:%d; time:%s", i, time.Now().Format(time.RFC3339Nano))
			log.Println(text)

			for i := 0; i < 2; i++ {
				key := fmt.Sprintf("key%d", i)
				m := &kafkago.Message{
					Key:   []byte(key),
					Value: []byte(text),
				}

				// write
				log.Println("write")
				require.NoError(t, w.WriteMessages(context.Background(), *m))

				defer func() {
					// change offset
					log.Println("commit")
					cm := kafkago.Message{
						Topic: Topic,
						Key:   []byte(key),
					}

					require.NoError(t, r.CommitMessages(context.Background(), cm))
				}()
			}

			// fetch
			log.Println("fetch")
			msg, err := r.FetchMessage(context.Background())
			require.NoError(t, err)
			//require.NoError(t, r.CommitMessages(context.Background(), msg))

			require.WithinDuration(t, time.Now(), msg.Time, time.Minute*2)
			require.True(t, msg.Offset >= 0)
			require.Equal(t,
				kafkago.Message{
					Topic:  Topic,
					Key:    []byte("key0"),
					Value:  []byte(text),
					Offset: msg.Offset,
					Time:   msg.Time,
				},
				msg)

			msg2, err := r.FetchMessage(context.Background())
			require.NoError(t, err)
			require.Equal(t, []byte("key1"), msg2.Key)
			require.Equal(t, msg.Offset+1, msg2.Offset)
		}()

	}
}
