package schemaregistry

import (
	"net/http"
	"time"
)

type IConfig interface {
	GetTimeout() time.Duration
	GetUrl() string
	GetTransport() (*http.Transport, error)
}
