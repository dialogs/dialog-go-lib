package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestZapLevel(t *testing.T) {

	require.Equal(t,
		zap.NewAtomicLevelAt(zap.DebugLevel),
		(&Config{Level: LevelDebug}).getZapLevel())

	require.Equal(t,
		zap.NewAtomicLevelAt(zap.ErrorLevel),
		(&Config{Level: LevelError}).getZapLevel())

	require.Equal(t,
		zap.NewAtomicLevelAt(zap.InfoLevel),
		(&Config{Level: LevelInfo}).getZapLevel())

	require.Equal(t,
		zap.NewAtomicLevelAt(zap.WarnLevel),
		(&Config{Level: LevelWarn}).getZapLevel())
}
