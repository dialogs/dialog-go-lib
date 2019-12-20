package logger

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestLoggerNew(t *testing.T) {

	for _, testInfo := range []struct {
		Args  []string
		Check func(*zap.Logger, func() string, string)
	}{
		{
			Args: []string{},
			Check: func(l *zap.Logger, getOut func() string, desc string) {
				require.True(t, l.Core().Enabled(zapcore.DebugLevel), desc, strings.Join(os.Args, " "))

				l.Debug("message#dbg")
				checkProdLoggerMessage(t, getOut(), "debug", "message#dbg")
			},
		},
		{
			Args: []string{_EnvLevel, zapcore.InfoLevel.String()},
			Check: func(l *zap.Logger, getOut func() string, desc string) {
				require.False(t, l.Core().Enabled(zapcore.DebugLevel), desc, strings.Join(os.Args, " "))
				require.True(t, l.Core().Enabled(zapcore.InfoLevel), desc, strings.Join(os.Args, " "))

				l.Info("message#info")
				checkProdLoggerMessage(t, getOut(), "info", "message#info")
			},
		},
		{
			Args: []string{_EnvLevel, zapcore.WarnLevel.String()},
			Check: func(l *zap.Logger, getOut func() string, desc string) {
				require.False(t, l.Core().Enabled(zapcore.InfoLevel), desc, strings.Join(os.Args, " "))
				require.True(t, l.Core().Enabled(zapcore.WarnLevel), desc, strings.Join(os.Args, " "))

				l.Warn("message#warn")
				checkProdLoggerMessage(t, getOut(), "warn", "message#warn")
			},
		},
		{
			Args: []string{_EnvLevel, zapcore.ErrorLevel.String(), _EnvTime, "iso8601"},
			Check: func(l *zap.Logger, getOut func() string, desc string) {
				require.False(t, l.Core().Enabled(zapcore.WarnLevel), desc, strings.Join(os.Args, " "))
				require.True(t, l.Core().Enabled(zapcore.ErrorLevel), desc, strings.Join(os.Args, " "))

				l.Error("message#err")
				checkProdLoggerMessage(t, getOut(), "error", "message#err")
			},
		},
		{
			Args: []string{_EnvLevel, zapcore.ErrorLevel.String(), _EnvTime, "iso8601"},
			Check: func(l *zap.Logger, getOut func() string, desc string) {
				require.False(t, l.Core().Enabled(zapcore.WarnLevel), desc, strings.Join(os.Args, " "))
				require.True(t, l.Core().Enabled(zapcore.ErrorLevel), desc, strings.Join(os.Args, " "))

				l.Error("message#err/iso8601")
				checkProdLoggerMessage(t, getOut(), "error", "message#err/iso8601")
			},
		},
		{
			Args: []string{_EnvLevel, zapcore.DebugLevel.String(), _EnvTime, "iso8601", _EnvMode, "1"},
			Check: func(l *zap.Logger, getOut func() string, desc string) {
				require.True(t, l.Core().Enabled(zapcore.DebugLevel), desc, strings.Join(os.Args, " "))

				l.Debug("message#debug/iso8601")
				checkDevLoggerMessage(t, getOut(), "debug", "message#debug/iso8601")
			},
		},
	} {

		func() {
			if len(testInfo.Args)%2 != 0 {
				t.Fatalf("invalid arguments list %v", testInfo.Args)
			}

			envNames := make([]string, 0)
			for i := 0; i < len(testInfo.Args); i += 2 {
				name := testInfo.Args[i]
				value := testInfo.Args[i+1]

				envNames = append(envNames, name)
				require.NoError(t, os.Setenv(name, value))
			}

			defer func() {
				for _, name := range envNames {
					require.NoError(t, os.Unsetenv(name))
				}
			}()

			r, w, err := os.Pipe()
			require.NoError(t, err)

			stdOut := os.Stdout
			stdErr := os.Stderr

			os.Stdout = w
			os.Stderr = w

			defer func() {
				os.Stdout = stdOut
				os.Stderr = stdErr
			}()

			fnGetOut := func() func() string {
				out := make(chan string, 1)

				go func() {
					buf := bytes.NewBuffer(nil)
					_, err := io.Copy(buf, r)
					require.NoError(t, err)
					out <- buf.String()
				}()

				return func() string {
					require.NoError(t, w.Close())
					return <-out
				}
			}

			desc := strings.Join(testInfo.Args, ",")

			l, err := New()
			require.NoError(t, err, desc)
			testInfo.Check(l, fnGetOut(), desc)
		}()

	}
}

func TestLoggerInvalidLevelName(t *testing.T) {

	require.NoError(t, os.Setenv(_EnvLevel, "warning"))
	defer func() { require.NoError(t, os.Unsetenv(_EnvLevel)) }()

	l, err := New()
	require.EqualError(t, err, `logger level (debug, info, warn, error, dpanic, panic, fatal): unrecognized level: "warning"`)
	require.Nil(t, l)
}

func checkProdLoggerMessage(t *testing.T, src, level, message string) {

	// source example:
	// "level":"debug","ts":"2019-10-15T17:21:10.215+0300","caller":"logger/logger_test.go:110","msg":"message#debug/iso8601"}

	msg := map[string]interface{}{}
	require.NoError(t, json.NewDecoder(bytes.NewReader([]byte(src))).Decode(&msg), src)

	require.Equal(t, level, msg["level"], src)
	require.Equal(t, message, msg["msg"], src)

	var (
		ts  time.Time
		err error
	)
	switch v := msg["ts"].(type) {
	case string:
		ts, err = time.Parse("2006-01-02T15:04:05.000Z0700", v)
	case float64:
		ts = time.Unix(int64(v), 0)
	default:
		t.Fatalf("unknown time type %#v", v)
	}

	require.NoError(t, err)
	require.WithinDuration(t, time.Now(), ts, time.Second)
}

func checkDevLoggerMessage(t *testing.T, src, level, message string) {

	// source example:
	// 2019-10-15T17:22:07.766+0300    DEBUG   logger/logger_test.go:110       message#debug/iso8601

	parts := strings.Split(src, "\t")
	require.Len(t, parts, 4)

	require.Equal(t, strings.ToUpper(level), parts[1], src)
	require.Equal(t, message+"\n", parts[3], src)

	ts, err := time.Parse("2006-01-02T15:04:05.000Z0700", parts[0])
	require.NoError(t, err)
	require.WithinDuration(t, time.Now(), ts, time.Second)
}
