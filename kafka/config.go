package kafka

import (
	"crypto/tls"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"strings"
	"time"
)

// Config for an implementation of kafka reader/writer
type Config struct {
	Brokers   []string
	TLSConfig *tls.Config
	Timeout   time.Duration
	DualStack bool
}

// Normalizes kafka config from dash-separated to dot-separated keys
func NormalizeConfig(in kafka.ConfigMap) kafka.ConfigMap {
	out := make(kafka.ConfigMap)
	for k, v := range in {
		out[strings.Trim(strings.Replace(k, "-", ".", -1), ".")] = v
	}
	return out
}