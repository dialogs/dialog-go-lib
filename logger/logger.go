package logger

import (
	"os"
	"strings"

	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New() (*zap.Logger, error) {

	const (
		FlagLevel       = "loglevel"
		FlagTime        = "logtime"
		FlagDevelopMode = "logdevelop"
	)

	cmdName := getCmdName(os.Args)
	flagSet := flag.NewFlagSet(cmdName, flag.ErrorHandling(-1)) // -1 - skip error 'not found flag'
	flagSet.SetOutput(newNopWriter())                           // don't print help information about unknown flags
	flagSet.ParseErrorsWhitelist.UnknownFlags = true            // skip unknown flags

	flagSet.StringVarP(new(string), FlagLevel, "", "", "logger level (debug, info, warn, error, dpanic, panic, fatal)")
	flagSet.StringVarP(new(string), FlagTime, "", "", "logger time format (iso8601, millis, nanos)")
	flagSet.BoolVarP(new(bool), FlagDevelopMode, "", false, "logger develop mode")

	if err := flagSet.ParseAll(os.Args, func(f *flag.Flag, v string) error {
		return f.Value.Set(strings.TrimSpace(v))
	}); err != nil {
		return nil, errors.Wrap(err, "failed to parse logger settings")
	}

	var (
		level                           = zapcore.DebugLevel
		timeFormat  zapcore.TimeEncoder = zapcore.EpochTimeEncoder
		developMode bool
	)

	{
		if f := flagSet.Lookup(FlagLevel); f != nil {
			if v := f.Value.String(); v != "" {
				if err := level.Set(v); err != nil {
					return nil, errors.Wrap(err, f.Usage)
				}
			}
		}

		if f := flagSet.Lookup(FlagTime); f != nil {
			if v := f.Value.String(); v != "" {
				if err := timeFormat.UnmarshalText([]byte(v)); err != nil {
					return nil, errors.Wrap(err, f.Usage)
				}
			}
		}

		if f := flagSet.Lookup(FlagDevelopMode); f != nil {
			if f.Value.String() == "true" {
				developMode = true
			}
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

func getCmdName(args []string) (cmdName string) {

	if len(args) > 1 {
		if v := args[1]; len(v) > 0 && v[0] != '-' {
			cmdName = v
		}
	}

	return
}
