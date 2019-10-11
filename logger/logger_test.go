package logger

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestLoggerNew(t *testing.T) {
	args := os.Args
	defer func() { os.Args = args }()

	argsClone := make([]string, len(args))
	copy(argsClone, args)

	for _, testInfo := range []struct {
		Args  []string
		Check func(*zap.Logger, string)
	}{
		{
			Args: []string{},
			Check: func(l *zap.Logger, desc string) {
				require.True(t, l.Core().Enabled(zapcore.DebugLevel), desc)
			},
		},
		{
			Args: append(argsClone, "--loglevel", zapcore.InfoLevel.String()),
			Check: func(l *zap.Logger, desc string) {
				require.False(t, l.Core().Enabled(zapcore.DebugLevel), desc)
				require.True(t, l.Core().Enabled(zapcore.InfoLevel), desc)
			},
		},
		{
			Args: append(argsClone, "--loglevel="+zapcore.WarnLevel.String()),
			Check: func(l *zap.Logger, desc string) {
				require.False(t, l.Core().Enabled(zapcore.InfoLevel), desc)
				require.True(t, l.Core().Enabled(zapcore.WarnLevel), desc)
			},
		},
		{
			Args: append(argsClone, "--loglevel="+zapcore.ErrorLevel.String(), "--logtime=iso8601"),
			Check: func(l *zap.Logger, desc string) {
				require.False(t, l.Core().Enabled(zapcore.WarnLevel), desc)
				require.True(t, l.Core().Enabled(zapcore.ErrorLevel), desc)
			},
		},
		{
			Args: append(argsClone[:1], "app", "--loglevel="+zapcore.ErrorLevel.String(), "--logtime", "iso8601"),
			Check: func(l *zap.Logger, desc string) {
				require.False(t, l.Core().Enabled(zapcore.WarnLevel), desc)
				require.True(t, l.Core().Enabled(zapcore.ErrorLevel), desc)
			},
		},
		{
			Args: append(argsClone[:1], "app", "-c", "config.yaml", "--loglevel="+zapcore.ErrorLevel.String(), "--logtime", "iso8601"),
			Check: func(l *zap.Logger, desc string) {
				require.False(t, l.Core().Enabled(zapcore.WarnLevel), desc)
				require.True(t, l.Core().Enabled(zapcore.ErrorLevel), desc)
			},
		},
		{
			Args: append(argsClone, "--loglevel="+zapcore.ErrorLevel.String(), "--logtime", "iso8601", "--logdevelop"),
			Check: func(l *zap.Logger, desc string) {
				require.False(t, l.Core().Enabled(zapcore.WarnLevel), desc)
				require.True(t, l.Core().Enabled(zapcore.ErrorLevel), desc)
			},
		},
	} {

		desc := strings.Join(testInfo.Args, ",")

		os.Args = testInfo.Args
		l, err := New()
		require.NoError(t, err, desc)
		testInfo.Check(l, desc)
	}

}

func TestLoggerInvalidLevelName(t *testing.T) {
	args := os.Args
	defer func() { os.Args = args }()

	os.Args = append(args, "--loglevel=warning")
	l, err := New()
	require.EqualError(t, err, `logger level (debug, info, warn, error, dpanic, panic, fatal): unrecognized level: "warning"`)
	require.Nil(t, l)
}
