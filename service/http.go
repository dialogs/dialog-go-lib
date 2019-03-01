package service

import (
	"context"
	"net/http"
	"time"
)

// A HTTP service
type HTTP struct {
	*service
	svr          *http.Server
	closeTimeout time.Duration
}

// NewHTTP create http service
func NewHTTP(addr string, handler http.Handler, closeTimeout time.Duration) *HTTP {
	return &HTTP{
		service: newService(),
		svr: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
		closeTimeout: closeTimeout,
	}
}

// Serve http
func (s *HTTP) Serve() error {

	run := func(retval chan<- error) {
		go func() {
			// check service state
			if err := PingConn(s.svr.Addr, serviceStartTimeout); err == nil {
				s.setReady(true)
			} else {
				s.Close()
			}
		}()

		retval <- s.svr.ListenAndServe()
	}

	stop := func() {
		ctx, cancel := context.WithTimeout(context.Background(), s.closeTimeout)
		defer cancel()

		s.svr.Shutdown(ctx)
	}

	return s.serve("http service", s.svr.Addr, run, stop)
}
