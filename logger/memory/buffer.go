package memory

import "bytes"

type Buffer struct {
	*bytes.Buffer
}

func NewBuffer() *Buffer {
	return &Buffer{
		Buffer: bytes.NewBuffer(nil),
	}
}

func (s *Buffer) Close() error { return nil }
func (s *Buffer) Sync() error  { return nil }
