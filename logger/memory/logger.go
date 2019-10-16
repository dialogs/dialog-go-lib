package memory

import (
	"net/url"
	"strconv"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(cfg *zap.Config) (*zap.Logger, *Buffer, error) {

	bufferID := "memory" + strconv.Itoa(int(time.Now().UnixNano()))

	buf := NewBuffer()
	err := zap.RegisterSink(bufferID, func(*url.URL) (zap.Sink, error) {
		return buf, nil
	})
	if err != nil {
		return nil, nil, err
	}

	if cfg == nil {
		prodConfig := zap.NewProductionConfig()
		cfg = &prodConfig
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}
	cfg.OutputPaths = append(cfg.OutputPaths, bufferID+"://")

	l, err := cfg.Build()
	return l, buf, err
}
