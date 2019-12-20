package schemaregistry

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"

	"github.com/pkg/errors"
)

func NewSchemaRegistryClient(cfg *Config) (*Client, error) {

	schemaregistryAddr := (&url.URL{
		Scheme: cfg.Scheme,
		Host:   net.JoinHostPort(cfg.Host, cfg.Port),
	}).String()

	var schemaregistryTransport *http.Transport
	if info, err := os.Stat(cfg.CA); err == nil {
		if info.Size() > (1024 * 1024 * 20) {
			return nil, errors.New("invalid CA file size")
		}

		pem, err := ioutil.ReadFile(cfg.CA)
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

	client, err := NewClient(schemaregistryAddr, cfg.Timeout, schemaregistryTransport)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create schema registry client")
	}

	return client, nil
}
