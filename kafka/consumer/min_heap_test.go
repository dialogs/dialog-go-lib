package consumer

import (
	"container/heap"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func messageForOffset(offset kafka.Offset) *kafka.Message {
	topic := "test"
	return &kafka.Message{TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: 1, Offset: offset}}
}

func TestMinHeap(t *testing.T) {
	slice := MessageHeap{}
	for _, v := range []int{2, 1, 5, 1, 3, 2, 1} {
		slice = append(slice, messageForOffset(kafka.Offset(v)))
	}
	h := &slice
	heap.Init(h)
	heap.Push(h, messageForOffset(8))
	min := kafka.Offset(0)
	for {
		if h.Len() == 0 {
			break
		}
		curr := heap.Pop(h).(kafka.Message)
		if curr.TopicPartition.Offset < min {
			t.Fatalf("Heap invariant violation. Got %d < %d", curr.TopicPartition.Offset, min)
		}
		min = curr.TopicPartition.Offset
	}
	if min != 8 {
		t.Fatalf("Maximum should be 8, got %d instead", min)
	}
}
