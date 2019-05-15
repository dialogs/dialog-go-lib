package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetString(t *testing.T) {

	// create environment
	os.Setenv("TEST_STRING_KEY1", "val1")
	os.Setenv("TEST_STRING_KEY2", "")

	v := New("test_string", true)

	{
		val, err := GetString(v, "key1")
		require.NoError(t, err)
		require.Equal(t, "val1", val)
	}

	{
		val, err := GetString(v, "key2")
		require.NoError(t, err)
		require.Equal(t, "", val)
	}

	{
		val, err := GetString(v, "key3")
		require.EqualError(t, err, "not found config value: 'key3'")
		require.Equal(t, "", val)
	}
}

func TestGetDuration(t *testing.T) {

	// create environment
	os.Setenv("TEST_DURATION_KEY1", "1")
	os.Setenv("TEST_DURATION_KEY2", "")

	v := New("test_duration", true)

	{
		val, err := GetDuration(v, "key1")
		require.NoError(t, err)
		require.Equal(t, time.Duration(1), val)
	}

	{
		val, err := GetDuration(v, "key2")
		require.NoError(t, err)
		require.Equal(t, time.Duration(0), val)
	}

	{
		val, err := GetDuration(v, "key3")
		require.EqualError(t, err, "not found config value: 'key3'")
		require.Equal(t, time.Duration(0), val)
	}
}

func TestGetBool(t *testing.T) {

	// create environment
	os.Setenv("TEST_BOOL_KEY1", "1")
	os.Setenv("TEST_BOOL_KEY2", "true")
	os.Setenv("TEST_BOOL_KEY3", "false")

	v := New("test_bool", true)

	{
		val, err := GetBool(v, "key1")
		require.NoError(t, err)
		require.Equal(t, true, val)
	}

	{
		val, err := GetBool(v, "key2")
		require.NoError(t, err)
		require.Equal(t, true, val)
	}

	{
		val, err := GetBool(v, "key3")
		require.NoError(t, err)
		require.Equal(t, false, val)
	}

	{
		val, err := GetBool(v, "key4")
		require.EqualError(t, err, "not found config value: 'key4'")
		require.Equal(t, false, val)
	}
}

func TestGetInt(t *testing.T) {

	// create environment
	os.Setenv("TEST_INT_KEY1", "1")
	os.Setenv("TEST_INT_KEY2", "")

	v := New("test_int", true)

	{
		val, err := GetInt(v, "key1")
		require.NoError(t, err)
		require.Equal(t, 1, val)
	}

	{
		val, err := GetInt(v, "key2")
		require.NoError(t, err)
		require.Equal(t, 0, val)
	}

	{
		val, err := GetInt(v, "key3")
		require.EqualError(t, err, "not found config value: 'key3'")
		require.Equal(t, 0, val)
	}
}

func TestGetFloat64(t *testing.T) {

	// create environment
	os.Setenv("TEST_F64_KEY1", "1")
	os.Setenv("TEST_F64_KEY2", "")

	v := New("test_f64", true)

	{
		val, err := GetFloat64(v, "key1")
		require.NoError(t, err)
		require.Equal(t, 1.0, val)
	}

	{
		val, err := GetFloat64(v, "key2")
		require.NoError(t, err)
		require.Equal(t, 0.0, val)
	}

	{
		val, err := GetFloat64(v, "key3")
		require.EqualError(t, err, "not found config value: 'key3'")
		require.Equal(t, 0.0, val)
	}
}
