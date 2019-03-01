package service

import (
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGRPC(t *testing.T) {

	h, p := tempAddress(t)
	service := NewGRPC(h + ":" + p)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		require.NoError(t, service.Serve())
	}()

	for !service.Ready() {
		// Wait
	}

	require.True(t, service.Ready())
	require.NoError(t, service.Close())
	wg.Wait()

	require.False(t, service.Ready())
}

func TestHTTP(t *testing.T) {

	h, p := tempAddress(t)
	svc := NewHTTP(h+":"+p, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), time.Second)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		require.Equal(t, http.ErrServerClosed, svc.Serve())
	}()

	for !svc.Ready() {
		// Wait
	}

	require.True(t, svc.Ready())
	require.NoError(t, svc.Close())
	wg.Wait()

	require.False(t, svc.Ready())
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
