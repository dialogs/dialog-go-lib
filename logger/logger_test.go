package logger

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestLogging(t *testing.T) {

	newout, err := ioutil.TempFile(os.TempDir(), "stdout")
	require.NoError(t, err)
	defer func() {
		name := newout.Name()
		defer func() {
			require.NoError(t, os.Remove(name))
		}()

		require.NoError(t, newout.Close())
	}()

	l, err := New(&Config{
		Debug:  false,
		Output: []string{newout.Name()},
	}, map[string]interface{}{
		"k1": "v1",
		"k2": "v2",
	})
	require.NoError(t, err)
	require.NotNil(t, l)

	l.Info("msg-inf", "val1")
	l.Error("msg-err", "val1")
	l.Warn("msg-wrn", "val1")
	l.Debug("msg-dbg", "val1")

	_, err = newout.Seek(0, io.SeekStart)
	require.NoError(t, err)

	var count int
	sc := bufio.NewScanner(newout)
	for sc.Scan() {
		m := make(map[string]interface{})
		r := strings.NewReader(sc.Text())
		require.NoError(t, json.NewDecoder(r).Decode(&m))

		require.Equal(t, "v1", m["k1"])
		delete(m, "k1")
		require.Equal(t, "v2", m["k2"])
		delete(m, "k2")

		require.NotEmpty(t, m["ts"])
		delete(m, "ts")

		switch count {
		case 0:
			require.Equal(t,
				map[string]interface{}{
					"level":   "info",
					"caller":  "logger/logger_test.go:40",
					"msg":     "msg-inf",
					"payload": `val1`,
				},
				m)
		case 1:
			require.NotEmpty(t, m["stacktrace"])
			delete(m, "stacktrace")

			require.Equal(t,
				map[string]interface{}{
					"level":   "error",
					"caller":  "logger/logger_test.go:41",
					"msg":     "msg-err",
					"payload": `val1`,
				},
				m)
		case 2:
			require.Equal(t,
				map[string]interface{}{
					"level":   "warn",
					"caller":  "logger/logger_test.go:42",
					"msg":     "msg-wrn",
					"payload": `val1`,
				},
				m)

		case 3:
			require.Equal(t,
				map[string]interface{}{
					"level":   "debug",
					"caller":  "logger/logger_test.go:43",
					"msg":     "msg-dbg",
					"payload": `val1`,
				},
				m)

		default:
			require.Fail(t, "unknown string")
		}
		count++
	}
	require.NoError(t, sc.Err())
}

func TestConvertValues(t *testing.T) {

	for res, src := range map[string][]interface{}{
		"1":          []interface{}{1},
		"1.2":        []interface{}{1.2},
		"err":        []interface{}{errors.New("err")},
		"{value}":    []interface{}{struct{ Value string }{"value"}},
		"str":        []interface{}{"str"},
		"1 true 1.5": []interface{}{1, true, 1.5},
	} {

		require.Equal(t,
			zap.String("payload", res),
			convertValues(src), res)
	}
}
