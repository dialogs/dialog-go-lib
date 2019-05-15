package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {

	// create environment
	os.Setenv("TEST_CONFIG_MAIN_KEY1", "val1")
	os.Setenv("TEST_CONFIG_MAIN_KEY2", "val2")
	os.Setenv("TEST_CONFIG_SUB_KEY1", "subval1")

	v := New("Test_Config_main", true)
	SetSub(v, New("test_config_sub", true), "sub")

	// test: check values
	require.Equal(t, "val1", v.GetString("key1"))
	require.Equal(t, "val2", v.GetString("key2"))
	require.Equal(t, "", v.GetString("key3"))
	require.Equal(t, "subval1", v.GetString("sub.key1"))

	// test: all settings
	require.Equal(t,
		map[string]interface{}{
			"key1": "val1",
			"key2": "val2",
			"sub": map[string]interface{}{
				"key1": "subval1",
			},
		},
		v.AllSettings())

	require.Equal(t,
		map[string]interface{}{
			"key1": "subval1",
		},
		v.Sub("sub").AllSettings())
}

func TestUnmarshal(t *testing.T) {

	// create environment
	os.Setenv("TEST_CONFIG_MAIN_KEY1", "val1")
	os.Setenv("TEST_CONFIG_MAIN_KEY2", "0")
	os.Setenv("TEST_CONFIG_SUB_KEY1", "1")

	v := New("Test_Config_main", true)
	SetSub(v, New("test_config_sub", true), "sub")

	type Sub struct {
		Key1 string `mapstructure:"key1"`
	}

	type Config struct {
		Key1 string        `mapstructure:"key1"`
		Key2 time.Duration `mapstructure:"key2"`
		Sub  Sub           `mapstructure:"sub"`
	}

	cfg := &Config{}
	require.NoError(t, v.Unmarshal(cfg))
	require.Equal(t,
		&Config{
			Key1: "val1",
			Key2: 0,
			Sub: Sub{
				Key1: "1",
			},
		},
		cfg)
}
