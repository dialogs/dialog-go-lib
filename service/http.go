package service

import (
	"context"
	"net/http"
	"time"
)

// A HTTP service
type HTTP struct {
	*service
	handler      http.Handler
	closeTimeout time.Duration
}

// NewHTTP create http service
func NewHTTP(handler http.Handler, closeTimeout time.Duration) *HTTP {
	return &HTTP{
		service:      newService(),
		handler:      handler,
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
	svr := http.Server{
		Addr:    addr,
		Handler: s.handler,
	}

	run := func(retval chan<- error) {
		retval <- svr.ListenAndServe()
	}

	stop := func() {
		ctx, cancel := context.WithTimeout(context.Background(), s.closeTimeout)
		defer cancel()

		svr.Shutdown(ctx)
	}

	return s.serve("http service", addr, run, stop)
}
