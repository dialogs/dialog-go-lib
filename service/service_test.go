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

	service := NewGRPC(grpc.WriteBufferSize(100))

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		require.NoError(t, service.ListenAndServeAddr(address))
	}()

	for !service.Ready() {
		// Wait
	}

	require.True(t, service.Ready())
	require.NoError(t, PingGRPC(address, time.Second))

	require.NoError(t, service.Close())
	wg.Wait()

	require.False(t, service.Ready())
	require.EqualError(t,
		PingGRPC(address, time.Microsecond),
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

	for !svc.Ready() {
		// Wait
	}

	require.True(t, svc.Ready())
	require.NoError(t, PingConn(address, time.Second))

	require.NoError(t, svc.Close())
	wg.Wait()

	require.False(t, svc.Ready())
	require.EqualError(t,
		PingConn(address, time.Microsecond),
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
