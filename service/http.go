package service

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/dialogs/dialog-go-lib/logger"
	pkgerr "github.com/pkg/errors"
)

// A HTTP service
type HTTP struct {
	svr          *http.Server
	closeTimeout time.Duration
	interrupt    chan os.Signal
	ready        int32
}

// NewHTTP create http service
func NewHTTP(addr string, handler http.Handler, closeTimeout time.Duration) *HTTP {
	return &HTTP{
		svr: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
		closeTimeout: closeTimeout,
		interrupt:    make(chan os.Signal),
	}
}

// Close service (io.Closer implementation)
func (s *HTTP) Close() (err error) {

	s.setReady(false)
	s.interrupt <- os.Interrupt

	return nil
}

// Serve http
func (s *HTTP) Serve() error {

	l, err := logger.New(&logger.Config{}, map[string]interface{}{
		"http service": s.svr.Addr,
	})
	if err != nil {
		return pkgerr.Wrap(err, "http service logger")
	}

	retval := make(chan error)
	defer close(retval)

	signal.Notify(s.interrupt, os.Interrupt, syscall.SIGTERM)

	go func() {
		retval <- s.svr.ListenAndServe()
	}()

	go func() {
		// check service state
		conn, err := net.DialTimeout("tcp", s.svr.Addr, time.Second*10)
		if err != nil {
			l.Error("Failed to start server", err)
		}
		defer conn.Close()
		s.setReady(true)
	}()

	l.Info("The service is ready to listen and serve")

	killSignal := <-s.interrupt
	switch killSignal {
	case os.Interrupt:
		l.Info("Got SIGINT...")
	case syscall.SIGTERM:
		l.Info("Got SIGTERM...")
	}

	l.Info("The service is shutting down", "timeout:", s.closeTimeout)

	ctx, cancel := context.WithTimeout(context.Background(), s.closeTimeout)
	defer cancel()

	l.Info("The service is shutting down...")
	s.svr.Shutdown(ctx)
	l.Info("The service is done")

	return <-retval
}

// Ready returns true if http service is available
func (s *HTTP) Ready() bool {
	return atomic.LoadInt32(&s.ready) > 0
}

func (s *HTTP) setReady(ok bool) {

	var val int32
	if ok {
		val = 1
	}

	atomic.StoreInt32(&s.ready, val)
}
