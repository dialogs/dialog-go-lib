package logger

import (
	"fmt"

	pkgerr "github.com/pkg/errors"
	"go.uber.org/zap"
)

// Logger wrapper
type Logger struct {
	logger  *zap.Logger
	context []zap.Field
}

// New logger wrapper
func New(cfg *Config, ctx map[string]interface{}) (*Logger, error) {

	zapCfg, err := cfg.toZapConfig()
	if err != nil {
		return nil, pkgerr.Wrap(err, "logger config")
	}

	if len(ctx) > 0 {
		if zapCfg.InitialFields == nil {
			zapCfg.InitialFields = map[string]interface{}{}
		}

		for k, v := range ctx {
			zapCfg.InitialFields[k] = v
		}
	}

	logger, err := zapCfg.Build(zap.AddCallerSkip(3))
	if err != nil {
		return nil, pkgerr.Wrap(err, "native logger")
	}

	context := make([]zap.Field, 0, len(ctx))

	return &Logger{
		logger:  logger,
		context: context,
	}, nil
}

// Info logs a message in info level
func (l *Logger) Info(msg string, value ...interface{}) {
	l.invoke(l.logger.Info, msg, value)
}

// Error logs a message in error level
func (l *Logger) Error(msg string, value ...interface{}) {
	l.invoke(l.logger.Error, msg, value)
}

// Debug logs a message in debug level
func (l *Logger) Debug(msg string, value ...interface{}) {
	l.invoke(l.logger.Debug, msg, value)
}

// Warn logs a message in warning level
func (l *Logger) Warn(msg string, value ...interface{}) {
	l.invoke(l.logger.Warn, msg, value)
}

// Fatal logs a message in fatal level
func (l *Logger) Fatal(msg string, value ...interface{}) {
	l.invoke(l.logger.Fatal, msg, value)
}

func (l *Logger) invoke(fn func(string, ...zap.Field), msg string, value []interface{}) {
	if len(value) == 0 {
		fn(msg)
	} else {
		fn(msg, convertValues(value))
	}
}

func convertValues(src []interface{}) zap.Field {

	// TODO: convert to zap types
	val := fmt.Sprintln(src...)
	if len(val) > 0 {
		val = val[:len(val)-1] // remove '\n'
	}

	return zap.String("payload", val)
}
