package service

import (
	"context"
	"errors"
	"net/http"
	"strings"
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
func (s *HTTP) ListenAndServeAddr(l *zap.Logger, addr string) error {
	s.SetAddr(addr)
	return s.ListenAndServe(l)
}

// ListenAndServe listens on the TCP network address and
// accepts incoming connections on the listener
func (s *HTTP) ListenAndServe(l *zap.Logger) error {

	var (
		svr  *http.Server
		addr = s.GetAddr()
	)

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

	run := func() error {
		if strings.TrimSpace(svr.Addr) == "" {
			return errors.New("invalid server address")
		}

		if svr.TLSConfig != nil {
			return svr.ListenAndServeTLS("", "")
		}

		return svr.ListenAndServe()
	}

	stop := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), s.closeTimeout)
		defer cancel()

		return svr.Shutdown(ctx)
	}

	return s.serve(l, "http service", addr, run, stop)
}
