package service

import (
	"net"

	"github.com/dialogs/dialog-go-lib/logger"
	pkgerr "github.com/pkg/errors"
	"google.golang.org/grpc"
)

// A GRPC service
type GRPC struct {
	*service
	svr *grpc.Server
}

// NewGRPC create grpc service
func NewGRPC(opts ...grpc.ServerOption) *GRPC {
	return &GRPC{
		service: newService(),
		svr:     grpc.NewServer(opts...),
	}
}

// RegisterService add new service to the grpc server
func (g *GRPC) RegisterService(fn func(svr *grpc.Server)) {
	fn(g.svr)
}

// ListenAndServeAddr listens on the TCP network address and
// accepts incoming connections on the listener
func (g *GRPC) ListenAndServeAddr(addr string) error {
	g.SetAddr(addr)
	return g.ListenAndServe()
}

// ListenAndServe listens on the TCP network address and
// accepts incoming connections on the listener
func (g *GRPC) ListenAndServe() error {

	addr := g.GetAddr()

	run := func(retval chan<- error) {

		l, err := net.Listen("tcp", addr)
		if err != nil {
			retval <- pkgerr.Wrap(err, "new grpc listener")
			return
		}

		retval <- g.svr.Serve(l)
	}

	stop := func(*logger.Logger) {
		g.svr.GracefulStop()
	}

	return g.serve("grpc service", addr, run, stop)
}
