package schemaregistry

import (
	"net/http"
	"time"
)

type ConfigMock struct {
	Url       string
	Timeout   time.Duration
	Transport *http.Transport
}

func NewConfigMock(url string, timeout time.Duration, transport *http.Transport) *ConfigMock {
	return &ConfigMock{
		Url:       url,
		Timeout:   timeout,
		Transport: transport,
	}
}

func (c *ConfigMock) GetUrl() string {
	return c.Url
}

func (c *ConfigMock) GetTimeout() time.Duration {
	return c.Timeout
}
func (c *ConfigMock) GetTransport() (*http.Transport, error) {
	return c.Transport, nil
}
