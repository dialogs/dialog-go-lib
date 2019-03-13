package consumer

import (
	"container/heap"
	"testing"

	"github.com/segmentio/kafka-go"
)

func messageForOffset(offset int64) kafka.Message {
	return kafka.Message{Topic: "test", Partition: 1, Offset: offset}
}

func TestMinHeap(t *testing.T) {
	slice := MessageHeap{}
	for _, v := range []int{2, 1, 5, 1, 3, 2, 1} {
		slice = append(slice, messageForOffset(int64(v)))
	}
	h := &slice
	heap.Init(h)
	heap.Push(h, messageForOffset(8))
	min := int64(0)
	for {
		if h.Len() == 0 {
			break
		}
		curr := heap.Pop(h).(kafka.Message)
		if curr.Offset < min {
			t.Fatalf("Heap invariant violation. Got %d < %d", curr.Offset, min)
		}
		min = curr.Offset
	}
	if min != 8 {
		t.Fatalf("Maximum should be 8, got %d instead", min)
	}
}
