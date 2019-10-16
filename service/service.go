package service

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	pkgerr "github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

func (s *service) serve(name, addr string, run func(retval chan<- error), stop func(*zap.Logger)) error {

	conf := zap.NewProductionConfig()
	conf.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	conf.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	l, err := conf.Build()
	if err != nil {
		return pkgerr.Wrap(err, "logger: "+name)
	}
	l = l.With(zap.String(name, addr))

	defer func() {
		l.Info("the service is done")
	}()

	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	retval := make(chan error)
	go run(retval)

	l.Info("the service is ready to listen and serve")

	go func() {
		select {
		case <-s.ctx.Done():
			l.Info("closing...")
			retval <- http.ErrServerClosed
		case killSignal := <-interrupt:
			switch killSignal {
			case os.Interrupt:
				l.Info("got SIGINT...")
			case syscall.SIGTERM:
				l.Info("got SIGTERM...")
			}
		}

		l.Info("the service is shutting down...")
		stop(l)
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
