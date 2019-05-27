package mocks

import (
	"bytes"
	"container/list"
	"sync"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

type buffer struct {
	data          *list.List
	offset        int64
	offsetVirtual int64
	mu            sync.RWMutex
}

func newBuffer() *buffer {
	return &buffer{
		data:          list.New(),
		offset:        -1,
		offsetVirtual: -1,
	}
}

func (b *buffer) Add(topic string, msg kafkago.Message) {

	msg.Time = time.Now()

	b.mu.Lock()
	b.data.PushBack(msg)
	b.mu.Unlock()
}

func (b *buffer) Get(topic string) *kafkago.Message {
	// TODO: wait messages

	b.mu.Lock()
	defer b.mu.Unlock()

	var index int64
	for e := b.data.Front(); e != nil; e = e.Next() {
		if b.offset < index && b.offsetVirtual < index {
			m := e.Value.(kafkago.Message)

			if topic == m.Topic {
				m.Offset = index
				b.offsetVirtual = index
				return &m
			}
		}

		index++
	}

	return nil
}

func (b *buffer) Commit(topic string, msg kafkago.Message) {
	b.mu.Lock()
	defer b.mu.Unlock()

	var index int64
	for e := b.data.Front(); e != nil; e = e.Next() {
		if b.offset < index {
			m := e.Value.(kafkago.Message)
			if topic == m.Topic && bytes.Equal(msg.Key, m.Key) {
				b.offset = index
				return
			}
		}

		index++
	}
}

func (b *buffer) Close() {
	b.mu.Lock()
	b.offsetVirtual = 0
	b.mu.Unlock()
}
