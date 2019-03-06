package service

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestGRPC(t *testing.T) {

	h, p := tempAddress(t)
	address := h + ":" + p

	svc := NewGRPC(grpc.WriteBufferSize(100))

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		require.NoError(t, svc.ListenAndServeAddr(address))
	}()

	clientOptions := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Second)}

	require.NoError(t, PingGRPC(address, 2, clientOptions...))

	require.Equal(t, address, svc.GetAddr())
	require.NoError(t, svc.Close())
	wg.Wait()

	require.EqualError(t,
		PingGRPC(address, 1, clientOptions...),
		"context deadline exceeded")
}

func TestHTTP(t *testing.T) {

	h, p := tempAddress(t)
	address := h + ":" + p

	svc := NewHTTP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), time.Second)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		require.Equal(t, http.ErrServerClosed, svc.ListenAndServeAddr(address))
	}()

	require.NoError(t, PingConn(address, 2, time.Second))

	require.Equal(t, address, svc.GetAddr())
	require.NoError(t, svc.Close())
	wg.Wait()

	require.EqualError(t,
		PingConn(svc.GetAddr(), 1, time.Microsecond),
		fmt.Sprintf("dial tcp %s: i/o timeout", address))
}

func tempAddress(t *testing.T) (host, port string) {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()

	host, port, err = net.SplitHostPort(l.Addr().String())
	require.NoError(t, err)
	return
}
