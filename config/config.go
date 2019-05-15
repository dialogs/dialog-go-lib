package config

import (
	"os"
	"strings"

	"github.com/spf13/viper"
)

// New returns a config with environment variables selected by name prefix
func New(prefix string, trimPrefix bool) *viper.Viper {

	v := viper.New()
	prefix = strings.ToUpper(prefix) + "_"

	for _, pair := range os.Environ() {
		if pos := strings.Index(pair, "="); pos != -1 {
			key := pair[:pos]
			if strings.HasPrefix(key, prefix) {
				newKey := key
				if trimPrefix {
					newKey = strings.TrimPrefix(newKey, prefix)
				}
				v.SetDefault(newKey, os.Getenv(key))
			}
		}
	}

	return v
}

// SetSub inserts map to config
func SetSub(dest, sub *viper.Viper, key string) {
	dest.Set(key, sub.AllSettings())
}
