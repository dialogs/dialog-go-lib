package logger

import (
	"os"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	_EnvLevel = "LOG_LEVEL"
	_EnvTime  = "LOG_TIME"
	_EnvMode  = "LOG_MODE"
)

func New() (*zap.Logger, error) {

	envLevel := os.Getenv(_EnvLevel)
	envTime := os.Getenv(_EnvTime)
	envMode := os.Getenv(_EnvMode)

	var (
		level                           = zapcore.DebugLevel
		timeFormat  zapcore.TimeEncoder = zapcore.EpochTimeEncoder
		developMode bool                = len(envMode) > 0
	)

	if envLevel != "" {
		if err := level.Set(envLevel); err != nil {
			return nil, errors.Wrap(err, "logger level (debug, info, warn, error, dpanic, panic, fatal)")
		}
	}

	if envTime != "" {
		if err := timeFormat.UnmarshalText([]byte(envTime)); err != nil {
			return nil, errors.Wrap(err, "logger time format (iso8601, millis, nanos)")
		}
	}

	cfg := zap.NewProductionConfig()
	if developMode {
		cfg = zap.NewDevelopmentConfig()
	}

	cfg.Level = zap.NewAtomicLevelAt(level)
	cfg.EncoderConfig.EncodeTime = timeFormat

	l, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return l, nil
}
