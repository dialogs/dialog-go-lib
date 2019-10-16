package memory

import (
	"encoding/json"
	"io"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestLogger(t *testing.T) {

	// two loggers for:
	// - zap.RegisterSink with unique name
	// - output buffer

	l1, buf1, err := New(nil)
	require.NoError(t, err)

	l2, buf2, err := New(nil)
	require.NoError(t, err)

	l1.Info("message for logger#1")
	l2.Debug("message for logger#2")

	checkLoggerMessage(t, buf1, "info", "message for logger#1", "memory/logger_test.go:"+getLine(3))
	checkLoggerMessage(t, buf2, "debug", "message for logger#2", "memory/logger_test.go:"+getLine(3))
}

func TestLoggerWithCustomConfig(t *testing.T) {

	customConf := zap.NewProductionConfig()
	customConf.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	customConf.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	l, buf, err := New(&customConf)
	require.NoError(t, err)

	l.Debug("message for logger#1")
	require.Equal(t, "", buf.String())

	l.Info("message for logger#2")
	checkLoggerMessage(t, buf, "info", "message for logger#2", "memory/logger_test.go:"+getLine(1))

	buf.Truncate(0)

	l.Warn("message for logger#3")
	checkLoggerMessage(t, buf, "warn", "message for logger#3", "memory/logger_test.go:"+getLine(1))
}

func checkLoggerMessage(t *testing.T, r io.Reader, level, message, caller string) {
	t.Helper()

	const (
		TagLevel  = "level"
		TagMsg    = "msg"
		TagCaller = "caller"
		TagTs     = "ts"
	)

	msg := map[string]string{}
	require.NoError(t, json.NewDecoder(r).Decode(&msg))

	require.Equal(t, level, msg[TagLevel])
	require.Equal(t, message, msg[TagMsg])
	require.Equal(t, caller, msg[TagCaller])

	ts, err := time.Parse("2006-01-02T15:04:05.000Z0700", msg[TagTs])
	require.NoError(t, err)
	require.WithinDuration(t, time.Now(), ts, time.Second)

	delete(msg, TagLevel)
	delete(msg, TagMsg)
	delete(msg, TagCaller)
	delete(msg, TagTs)
	require.Len(t, msg, 0)
}

func getLine(diff int) string {
	_, _, line, _ := runtime.Caller(1)
	return strconv.Itoa(line - diff)
}
