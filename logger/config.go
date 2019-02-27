package logger

import (
	"os"

	pkgerr "github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Level of the writer
type Level int32

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// Config of the logger wrapper
type Config struct {
	Debug  bool     `json:"debug" mapstructure:"debug"`
	Level  Level    `json:"level" mapstructure:"level"`
	Output []string `json:"output" mapstructure:"output"`
}

func (c Config) toZapConfig() (*zap.Config, error) {

	var zapCfg zap.Config
	if c.Debug {
		zapCfg = zap.NewDevelopmentConfig()
	} else {
		zapCfg = zap.NewProductionConfig()
	}

	if len(c.Output) > 0 {
		for _, path := range c.Output {
			if _, err := os.Stat(path); err != nil {
				return nil, pkgerr.Wrap(err, "logger output path")
			}
			zapCfg.OutputPaths = append(zapCfg.OutputPaths, path)
		}
	}

	zapCfg.Level = c.getZapLevel()

	return &zapCfg, nil
}

func (c Config) getZapLevel() zap.AtomicLevel {

	switch c.Level {
	case LevelDebug:
		return zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case LevelInfo:
		return zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case LevelWarn:
		return zap.NewAtomicLevelAt(zapcore.WarnLevel)
	}

	return zap.NewAtomicLevelAt(zapcore.ErrorLevel)
}
