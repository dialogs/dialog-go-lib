package consumer

import "github.com/confluentinc/confluent-kafka-go/kafka"

type MessageHeap []*kafka.Message

func (h MessageHeap) Len() int { return len(h) }
func (h MessageHeap) Less(i, j int) bool {
	return h[i].TopicPartition.Offset < h[j].TopicPartition.Offset
}
func (h MessageHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *MessageHeap) Push(x interface{}) {
	*h = append(*h, x.(*kafka.Message))
}

func (h *MessageHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
