package service

import (
	"net"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/dialogs/dialog-go-lib/logger"
	pkgerr "github.com/pkg/errors"
	"google.golang.org/grpc"
)

const serviceStartTimeout = time.Second * 15

// service base object
type service struct {
	interrupt chan os.Signal
	ready     int32
}

// create base service object
func newService() *service {
	return &service{
		interrupt: make(chan os.Signal),
	}
}

// Close service (io.Closer implementation)
func (s *service) Close() (err error) {

	s.setReady(false)
	s.interrupt <- os.Interrupt

	return nil
}

// Ready returns true if service is available
func (s *service) Ready() bool {
	return atomic.LoadInt32(&s.ready) > 0
}

func (s *service) setReady(ok bool) {

	var val int32
	if ok {
		val = 1
	}

	atomic.StoreInt32(&s.ready, val)
}

func (s *service) serve(name, addr string, run func(retval chan<- error), stop func()) error {

	l, err := logger.New(&logger.Config{}, map[string]interface{}{
		name: addr,
	})
	if err != nil {
		return pkgerr.Wrap(err, "logger: "+name)
	}

	signal.Notify(s.interrupt, os.Interrupt, syscall.SIGTERM)

	retval := make(chan error)
	go run(retval)

	l.Info("The service is ready to listen and serve")

	killSignal := <-s.interrupt
	switch killSignal {
	case os.Interrupt:
		l.Info("Got SIGINT...")
	case syscall.SIGTERM:
		l.Info("Got SIGTERM...")
	}

	l.Info("The service is shutting down...")
	stop()
	l.Info("The service is done")

	return <-retval
}

// PingConn ping tcp connection
func PingConn(addr string, timeout time.Duration) error {

	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return err
	}
	defer conn.Close()

	return nil
}

// PingGRPC ping grpc server
func PingGRPC(addr string, timeout time.Duration) error {

	conn, err := grpc.Dial(addr,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(timeout))
	if err != nil {
		return err
	}
	defer conn.Close()

	return nil
}
