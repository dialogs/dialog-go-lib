package db

import (
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Config struct {
	Host                   string            `mapstructure:"host"`
	Port                   string            `mapstructure:"port"`
	Name                   string            `mapstructure:"name"`
	User                   string            `mapstructure:"user"`
	Password               string            `mapstructure:"password"`
	SslMode                string            `mapstructure:"ssl-mode"`
	Scheme                 string            `mapstructure:"scheme"`
	HealthCheckPeriod      time.Duration     `mapstructure:"health-check"`
	MaxConnections         int               `mapstructure:"max-connections"`
	StatementCacheCapacity int               `mapstructure:"statement-cache-capacity"`
	RawProperties          map[string]string `mapstructure:"raw-properties"`
}

func NewConfig(src *viper.Viper) (conf Config, err error) {

	if err = src.Unmarshal(&conf); err != nil {
		err = errors.Wrap(err, "failed to parse db config")
	}

	return
}

func (c Config) Check() error {

	if len(c.Host) == 0 {
		return errors.New("db.host: was not set")
	}

	if len(c.Port) == 0 {
		return errors.New("db.port: invalid value")
	}

	if len(c.Name) == 0 {
		return errors.New("db.name: was not set")
	}

	if val := strings.TrimSpace(c.Scheme); val != c.Scheme {
		return errors.New("db.scheme: invalid value")
	}

	if len(c.User) == 0 {
		return errors.New("db.user: was not set")
	}

	if len(c.Password) == 0 {
		return errors.New("db.password: was not set")
	}

	if len(c.SslMode) == 0 {
		return errors.New("db.ssl-mode: was not set")
	}

	if c.HealthCheckPeriod < 0 {
		return errors.New("db.health-check: invalid value")
	}

	if c.MaxConnections < 0 {
		return errors.New("db.max-connections' invalid value")
	}

	if c.StatementCacheCapacity < 0 {
		return errors.New("db.statement-cache-capacity' invalid value")
	}

	return nil
}

func (c *Config) Addr() string {
	return net.JoinHostPort(c.Host, c.Port)
}

func (c *Config) PoolConnURL() string {
	return c.URL(true, true).String()
}

func (c *Config) ConnURL() string {
	return c.URL(false, true).String()
}

func (c *Config) ConnURLWithoutSchema() string {
	return c.URL(false, false).String()
}

func (c *Config) URL(poll, withScheme bool) *url.URL {

	q := url.Values{}
	q.Set("sslmode", c.SslMode)
	if c.Scheme != "" && withScheme {
		q.Set("search_path", c.Scheme)
	}

	if poll {
		q.Set("statement_cache_mode", "describe")

		// in the pgx library exist default value: 512
		if c.StatementCacheCapacity > 0 {
			q.Set("statement_cache_capacity", strconv.Itoa(c.StatementCacheCapacity))
		}

		// in the pgx library exist default value: 4
		if c.MaxConnections > 0 {
			q.Set("pool_max_conns", strconv.Itoa(c.MaxConnections))
		}

		// in the pgx library exist default value: minute
		if c.HealthCheckPeriod > 0 {
			q.Set("pool_health_check_period", c.HealthCheckPeriod.String())
		}

		if len(c.RawProperties) > 0 {
			for k, v := range c.RawProperties {
				q.Set(k, v)
			}
		}
	}

	return &url.URL{
		Scheme:   "postgres",
		Host:     c.Addr(),
		User:     url.UserPassword(c.User, c.Password),
		Path:     c.Name,
		RawQuery: q.Encode(),
	}
}
