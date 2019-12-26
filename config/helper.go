package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// GetString returns string value. Returns error if value is not set
func GetString(src *viper.Viper, key string) (string, error) {

	if src.IsSet(key) {
		return src.GetString(key), nil
	}

	return "", newError(key)
}

// GetDuration returns duration value. Returns error if value is not set
func GetDuration(src *viper.Viper, key string) (time.Duration, error) {

	if src.IsSet(key) {
		return src.GetDuration(key), nil
	}

	return 0, newError(key)
}

// GetBool returns boolean value. Returns error if value is not set
func GetBool(src *viper.Viper, key string) (bool, error) {

	if src.IsSet(key) {
		return src.GetBool(key), nil
	}

	return false, newError(key)
}

// GetInt returns integer value. Returns error if value is not set
func GetInt(src *viper.Viper, key string) (int, error) {

	if src.IsSet(key) {
		return src.GetInt(key), nil
	}

	return 0, newError(key)
}

// GetFloat64 returns float64 value. Returns error if value is not set
func GetFloat64(src *viper.Viper, key string) (float64, error) {

	if src.IsSet(key) {
		return src.GetFloat64(key), nil
	}

	return 0, newError(key)
}

func NewCobraInitializer(cfgFile *string) func() {
	return func() {
		if strings.TrimSpace(*cfgFile) == "" {
			log.Fatal("invalid config file name")
		} else if _, err := os.Stat(*cfgFile); err != nil {
			log.Fatal(err)
		}

		// Use config file from the flag.
		viper.SetConfigFile(*cfgFile)
		viper.AutomaticEnv() // read in environment variables that match

		// If a config file is found, read it in.
		if err := viper.ReadInConfig(); err != nil {
			log.Fatal("failed to parse config:", err)
		}

		log.Println("using config file:", viper.ConfigFileUsed())
	}
}

func newError(key string) error {
	return fmt.Errorf("not found config value: '%s'", key)
}
