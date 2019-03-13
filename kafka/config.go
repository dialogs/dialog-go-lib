package kafka

import (
	"crypto/tls"
	"time"
)

// Config for an implementation of kafka reader/writer
type Config struct {
	Brokers   []string
	GroupID   string
	TLSConfig *tls.Config
	Timeout   time.Duration
	DualStack bool
}
