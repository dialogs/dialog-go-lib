package service

import (
	"context"
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// A HTTP service
type HTTP struct {
	*service
	closeTimeout time.Duration

	// use only once property
	handler http.Handler
	server  *http.Server
}

// NewHTTP creates a http service with the handler
func NewHTTP(handler http.Handler, closeTimeout time.Duration) *HTTP {
	return &HTTP{
		service:      newService(),
		handler:      handler,
		closeTimeout: closeTimeout,
	}
}

// NewHTTPWithServer creates a http service with a custom http server configuration
func NewHTTPWithServer(server *http.Server, closeTimeout time.Duration) *HTTP {
	return &HTTP{
		service:      newService(),
		server:       server,
		closeTimeout: closeTimeout,
	}
}

// ListenAndServeAddr listens on the TCP network address and
// accepts incoming connections on the listener
func (s *HTTP) ListenAndServeAddr(addr string) error {
	s.SetAddr(addr)
	return s.ListenAndServe()
}

// ListenAndServe listens on the TCP network address and
// accepts incoming connections on the listener
func (s *HTTP) ListenAndServe() error {

	addr := s.GetAddr()

	var svr *http.Server
	if s.server != nil {
		svr = s.server
		if addr != "" {
			svr.Addr = addr
		}

	} else {
		svr = &http.Server{
			Addr:    addr,
			Handler: s.handler,
		}

	}

	run := func(retval chan<- error) {
		if svr.Addr == "" {
			retval <- errors.New("invalid server address")
			return
		}

		if svr.TLSConfig != nil {
			// use svr.TLSConfig.Certificates for certificates
			retval <- svr.ListenAndServeTLS("", "")
		} else {
			retval <- svr.ListenAndServe()
		}
	}

	stop := func(l *zap.Logger) {
		ctx, cancel := context.WithTimeout(context.Background(), s.closeTimeout)
		defer cancel()

		if err := svr.Shutdown(ctx); err != nil {
			l.Error("failed to shutdown", zap.Error(err))
		}
	}

	return s.serve("http service", addr, run, stop)
}
