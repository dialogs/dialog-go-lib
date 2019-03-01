package service

import (
	"net"

	pkgerr "github.com/pkg/errors"
	"google.golang.org/grpc"
)

// A GRPC service
type GRPC struct {
	*service
	addr string
	svr  *grpc.Server
}

// NewGRPC create grpc service
func NewGRPC(addr string, opts ...grpc.ServerOption) *GRPC {
	return &GRPC{
		service: newService(),
		addr:    addr,
		svr:     grpc.NewServer(opts...),
	}
}

// RegisterService add new service to the grpc server
func (g *GRPC) RegisterService(fn func(svr *grpc.Server)) {
	fn(g.svr)
}

// Serve grpc
func (g *GRPC) Serve() error {

	run := func(retval chan<- error) {
		go func() {
			if err := PingGRPC(g.addr, serviceStartTimeout); err == nil {
				g.setReady(true)
			} else {
				g.Close()
			}
		}()

		l, err := net.Listen("tcp", g.addr)
		if err != nil {
			retval <- pkgerr.Wrap(err, "new grpc listener")
			return
		}

		retval <- g.svr.Serve(l)
	}

	stop := func() {
		g.svr.GracefulStop()
	}

	return g.serve("grpc service", g.addr, run, stop)
}
