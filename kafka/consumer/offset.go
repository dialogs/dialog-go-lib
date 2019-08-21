package consumer

import (
	"sync"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type offset struct {
	topics  map[string]map[int32]kafka.Offset
	counter int
	mu      sync.RWMutex
}

func newOffset() *offset {
	return &offset{
		topics: make(map[string]map[int32]kafka.Offset),
	}
}

func (o *offset) Add(in ...kafka.TopicPartition) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.counter += len(in)

	for i := range in {
		val := &in[i]

		var topic string
		if val.Topic != nil {
			topic = *val.Topic
		}

		partitons, ok := o.topics[topic]
		if ok {
			partitonOffset, ok := partitons[val.Partition]
			if !ok || partitonOffset < val.Offset {
				partitons[val.Partition] = val.Offset
			}

		} else {
			partitons = map[int32]kafka.Offset{
				val.Partition: val.Offset,
			}
		}

		o.topics[topic] = partitons
	}
}

func (o *offset) Sync(in ...kafka.TopicPartition) {

	inMap := partitionsListToMap(in)
	o.mu.Lock()
	defer o.mu.Unlock()

	{
		// fix topics list
		removeTopics := make([]string, 0, len(o.topics))
		for topic := range o.topics {
			if _, ok := inMap[topic]; !ok {
				removeTopics = append(removeTopics, topic)
			}
		}

		for _, topic := range removeTopics {
			delete(o.topics, topic)
		}
	}

	{
		// fix partitions
		for topic, srcPartitions := range o.topics {
			removePartitions := make([]int32, 0, len(srcPartitions))

			newPartitions := inMap[topic]
			for num := range srcPartitions {
				if _, ok := newPartitions[num]; !ok {
					removePartitions = append(removePartitions, num)
				}
			}

			for _, num := range removePartitions {
				delete(srcPartitions, num)
			}
		}
	}
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
		delete(partitions, in.Partition)
		if len(partitions) == 0 {
			delete(o.topics, topic)
		}

		o.counter--
	}
	o.mu.Unlock()
}

func (o *offset) Get() *kafka.TopicPartition {
	o.mu.Lock()
	defer o.mu.Unlock()

	for topic, partition := range o.topics {
		for p, po := range partition {
			return &kafka.TopicPartition{
				Topic:     stringPointer(topic),
				Partition: p,
				Offset:    po,
			}
		}
	}

	return nil
}

func (o *offset) GetOffset(topicPtr *string, num int32) (retval kafka.Offset) {

	retval = -1

	var topic string
	if topicPtr != nil {
		topic = *topicPtr
	}

	o.mu.RLock()
	if p, ok := o.topics[topic]; ok {
		if val, ok := p[num]; ok {
			retval = val
		}
	}
	o.mu.RUnlock()

	return
}

func partitionsListToMap(in []kafka.TopicPartition) map[string]map[int32]kafka.Offset {

	retval := make(map[string]map[int32]kafka.Offset)

	for i := range in {
		val := &in[i]

		var topic string
		if val.Topic != nil {
			topic = *val.Topic
		}

		partitons, ok := retval[topic]
		if !ok {
			partitons = map[int32]kafka.Offset{}
		}

		partitons[val.Partition] = val.Offset

		retval[topic] = partitons
	}

	return retval
}
