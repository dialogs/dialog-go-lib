package schemaregistry

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/pkg/errors"
)

type Config struct {
	Scheme  string        `mapstructure:"scheme"`
	Host    string        `mapstructure:"host"`
	Port    string        `mapstructure:"port"`
	Timeout time.Duration `mapstructure:"timeout"`
	CA      string        `mapstructure:"ca"` //  path to pem file
}

func (c *Config) GetTimeout() time.Duration {
	return c.Timeout
}

func (c *Config) GetUrl() string {
	return (&url.URL{
		Scheme: c.Scheme,
		Host:   net.JoinHostPort(c.Host, c.Port),
	}).String()
}

func (c *Config) GetTransport() (*http.Transport, error) {
	var schemaregistryTransport *http.Transport
	if info, err := os.Stat(c.CA); err == nil {
		if info.Size() > (1024 * 1024 * 20) {
			return nil, errors.New("invalid CA file size")
		}

		pem, err := ioutil.ReadFile(c.CA)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read schema registry CA file")
		}

		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return nil, errors.New("failed to parse schema registry CA")
		}

		schemaregistryTransport = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: pool,
			},
		}
	}
	return schemaregistryTransport, nil
}
