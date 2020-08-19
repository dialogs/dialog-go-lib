package service

import (
	"net"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// A GRPC service
type GRPC struct {
	*service
	svr          *grpc.Server
	closeTimeout time.Duration
}

// NewGRPC create grpc service
func NewGRPC(opts ...grpc.ServerOption) *GRPC {
	return &GRPC{
		service:      newService(),
		svr:          grpc.NewServer(opts...),
		closeTimeout: time.Second * 30,
	}
}

func (g *GRPC) WithCloseTimeout(val time.Duration) *GRPC {
	g.closeTimeout = val
	return g
}

// RegisterService add new service to the grpc server
func (g *GRPC) RegisterService(fn func(svr *grpc.Server)) {
	fn(g.svr)
}

// ListenAndServeAddr listens on the TCP network address and
// accepts incoming connections on the listener
func (g *GRPC) ListenAndServeAddr(l *zap.Logger, addr string) error {
	g.SetAddr(addr)
	return g.ListenAndServe(l)
}

// ListenAndServe listens on the TCP network address and
// accepts incoming connections on the listener
func (g *GRPC) ListenAndServe(l *zap.Logger) error {

	addr := g.GetAddr()

	run := func() error {
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			return err
		}

		return g.svr.Serve(listener)
	}

	stop := func() error {

		stopped := make(chan struct{})
		go func() {
			g.svr.GracefulStop()
			close(stopped)
		}()

		tm := time.NewTimer(g.closeTimeout)
		select {
		case <-tm.C:
			l.Warn("closed by timeout")
			g.svr.Stop()
		case <-stopped:
			tm.Stop()
		}

		return nil
	}

	return g.serve(l, "grpc service", addr, run, stop)
}
