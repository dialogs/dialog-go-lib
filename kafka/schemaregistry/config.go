package schemaregistry

import "time"

type Config struct {
	Scheme  string        `mapstructure:"scheme"`
	Host    string        `mapstructure:"host"`
	Port    string        `mapstructure:"port"`
	Timeout time.Duration `mapstructure:"timeout"`
	CA      string        `mapstructure:"ca"` //  path to pem file
}
