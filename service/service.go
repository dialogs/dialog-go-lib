package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// service base object
type service struct {
	addr      string
	mu        sync.RWMutex
	ctx       context.Context
	ctxCancel func()
}

// create base service object
func newService() *service {

	ctx, cancel := context.WithCancel(context.Background())

	return &service{
		ctx:       ctx,
		ctxCancel: cancel,
	}
}

// Close service (io.Closer implementation)
func (s *service) Close() (err error) {
	s.ctxCancel()
	return nil
}

// SetAddr set listener address
func (s *service) SetAddr(val string) {

	s.mu.Lock()
	s.addr = val
	s.mu.Unlock()
}

// GetAddr returns listener address
func (s *service) GetAddr() (addr string) {

	s.mu.RLock()
	addr = s.addr
	s.mu.RUnlock()

	return
}

func (s *service) serve(l *zap.Logger, name, addr string, run, stop func() error) error {

	select {
	case <-s.ctx.Done():
		return http.ErrServerClosed
	default:
		// nothing do
	}

	if l == nil {
		l = zap.NewNop()
	}

	l = l.With(zap.String(name, addr))
	defer func() {
		s.ctxCancel()
		l.Info("the service is done")
	}()

	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	retval := make(chan error)
	go func() { retval <- run() }()

	l.Info("the service is ready to listen and serve")

	go func() {
		select {
		case <-s.ctx.Done():
			l.Info("closing...")
		case killSignal := <-interrupt:
			l.Info(fmt.Sprintf("got %s...", killSignal.String()))
		}

		l.Info("the service is shutting down...")
		if err := stop(); err != nil {
			l.Error("failed to shutdown", zap.Error(err))
		}
	}()

	return <-retval
}

// PingConn ping tcp connection
func PingConn(addr string, tries int, timeout time.Duration, tlsConf *tls.Config) (err error) {

	for i := 0; i < tries; i++ {
		var conn io.Closer
		if tlsConf != nil {
			dialer := &net.Dialer{KeepAlive: timeout}
			conn, err = tls.DialWithDialer(dialer, "tcp", addr, tlsConf)
		} else {
			conn, err = net.DialTimeout("tcp", addr, timeout)
		}
		if err == nil {
			defer conn.Close()
			return
		}

		if isLast := i == tries-1; !isLast {
			time.Sleep(time.Second)
		}
	}

	return
}

// PingGRPC ping grpc server
func PingGRPC(addr string, tries int, opts ...grpc.DialOption) (err error) {

	for i := 0; i < tries; i++ {
		var conn *grpc.ClientConn
		conn, err = grpc.Dial(addr, opts...)
		if err == nil {
			defer conn.Close()
			return
		}

		if isLast := i == tries-1; !isLast {
			time.Sleep(time.Second)
		}
	}

	return
}
