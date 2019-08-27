package consumer

import (
	"sync"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type offsetEntry struct {
	Offset kafka.Offset
	Count  int
}

type offset struct {
	topics  map[string]map[int32]*offsetEntry
	counter int
	mu      sync.RWMutex
}

func newOffset() *offset {
	return &offset{
		topics: make(map[string]map[int32]*offsetEntry),
	}
}

func (o *offset) Add(in ...kafka.TopicPartition) {
	o.mu.Lock()
	defer o.mu.Unlock()

	for i := range in {
		val := &in[i]

		var topic string
		if val.Topic != nil {
			topic = *val.Topic
		}

		partitons, ok := o.topics[topic]
		if ok {
			partitonOffset, ok := partitons[val.Partition]
			if !ok {
				partitonOffset = &offsetEntry{}
				partitons[val.Partition] = partitonOffset
			}

			if !ok || partitonOffset.Offset < val.Offset {
				partitonOffset.Offset = val.Offset
				partitonOffset.Count++
				o.counter++
			}

		} else {
			partitons = map[int32]*offsetEntry{
				val.Partition: &offsetEntry{Offset: val.Offset, Count: 1},
			}
			o.counter++
		}

		o.topics[topic] = partitons
	}
}

func (o *offset) Clear() {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.topics = make(map[string]map[int32]*offsetEntry)
	o.counter = 0
}

func (o *offset) Counter() (counter int) {
	o.mu.RLock()
	counter = o.counter
	o.mu.RUnlock()

	return
}

func (o *offset) Remove(in kafka.TopicPartition) {

	var topic string
	if in.Topic != nil {
		topic = *in.Topic
	}

	o.mu.Lock()

	partitions, ok := o.topics[topic]
	if ok {
		entry := partitions[in.Partition]
		delete(partitions, in.Partition)
		if len(partitions) == 0 {
			delete(o.topics, topic)
		}

		o.counter -= entry.Count
	}

	if len(o.topics) == 0 {
		o.counter = 0
	}

	o.mu.Unlock()
}

func (o *offset) Get() (retval []kafka.TopicPartition) {

	o.mu.RLock()
	for topic, partition := range o.topics {
		for p, po := range partition {
			retval = append(retval, kafka.TopicPartition{
				Topic:     stringPointer(topic),
				Partition: p,
				Offset:    po.Offset,
			})
		}
	}
	o.mu.RUnlock()

	return
}
