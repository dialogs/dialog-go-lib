package service

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/dialogs/dialog-go-lib/cert"
	"github.com/dialogs/dialog-go-lib/logger"
	"github.com/dialogs/dialog-go-lib/service/test"
	"github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type brokenChecker struct {
}

func (c *brokenChecker) Ping(context.Context, *types.Empty) (*types.Empty, error) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
	return &types.Empty{}, nil
}

func init() {
	// for debug
	log.SetFlags(log.Llongfile | log.Ltime | log.Lmicroseconds)
}

func TestGRPCShutdownWithTimeout(t *testing.T) {

	l, err := logger.New()
	require.NoError(t, err)

	h, p := tempAddress(t)
	address := net.JoinHostPort(h, p)

	svc := NewGRPC()
	require.Equal(t, time.Second*30, svc.closeTimeout)
	svc.WithCloseTimeout(time.Second)
	require.Equal(t, time.Second, svc.closeTimeout)

	chCloseTimestamps := make(chan time.Time, 1)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		svc.RegisterService(func(g *grpc.Server) {
			test.RegisterCheckerServer(g, &brokenChecker{})
		})
		require.NoError(t, svc.ListenAndServeAddr(l, address))

		start := <-chCloseTimestamps
		sub := time.Since(start)
		require.True(t, sub >= time.Second, sub.String())
		require.True(t, sub < time.Millisecond*1500, sub.String())
	}()

	clientOptions := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Second)}

	require.NoError(t, PingGRPC(address, 2, clientOptions...))

	clientConn, err := grpc.Dial(address, clientOptions...)
	require.NoError(t, err)

	client := test.NewCheckerClient(clientConn)

	go func() {
		time.Sleep(time.Microsecond * 200)
		chCloseTimestamps <- time.Now()
		require.NoError(t, svc.Close())
	}()

	_, err = client.Ping(context.Background(), &types.Empty{})
	require.EqualError(t, err, "rpc error: code = Unavailable desc = transport is closing")
}

func TestGRPC(t *testing.T) {

	l, err := logger.New()
	require.NoError(t, err)

	h, p := tempAddress(t)
	address := net.JoinHostPort(h, p)

	svc := NewGRPC(grpc.WriteBufferSize(100))

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		require.NoError(t, svc.ListenAndServeAddr(l, address))

		// test: safe close
		require.NoError(t, svc.Close())
	}()

	clientOptions := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Second)}

	require.NoError(t, PingGRPC(address, 2, clientOptions...))

	require.Equal(t, address, svc.GetAddr())
	require.NoError(t, syscall.Kill(syscall.Getpid(), syscall.SIGINT))
	wg.Wait()

	require.Equal(t,
		context.DeadlineExceeded,
		PingGRPC(address, 1, clientOptions...))
}

func TestHTTP(t *testing.T) {

	l, err := logger.New()
	require.NoError(t, err)

	h, p := tempAddress(t)
	address := net.JoinHostPort(h, p)

	svc := NewHTTP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), time.Second)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		require.Equal(t, http.ErrServerClosed, svc.ListenAndServeAddr(l, address))

		// test: safe close
		require.NoError(t, svc.Close())
	}()

	require.NoError(t, PingConn(address, 2, time.Second, nil))

	require.Equal(t, address, svc.GetAddr())
	require.NoError(t, syscall.Kill(syscall.Getpid(), syscall.SIGTERM))
	wg.Wait()

	err = PingConn(svc.GetAddr(), 1, time.Microsecond, nil)
	eNet := err.(*net.OpError)
	require.Equal(t, eNet.Op, "dial")
	require.True(t, eNet.Timeout())
	require.Equal(t, address, eNet.Addr.String())
}

func TestHTTPCloseBeforeRun(t *testing.T) {

	h, p := tempAddress(t)
	address := net.JoinHostPort(h, p)

	svc := NewHTTP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), time.Second)

	require.NoError(t, svc.Close())

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		require.Equal(t, http.ErrServerClosed, svc.ListenAndServeAddr(nil, address))

		// test: safe close
		require.NoError(t, svc.Close())
	}()

	{
		err := PingConn(address, 2, time.Second, nil)
		eNet := err.(*net.OpError)
		require.Equal(t, eNet.Op, "dial")
		eSys := eNet.Err.(*os.SyscallError)
		require.EqualError(t, eSys.Err, "connection refused")
	}

	require.Equal(t, address, svc.GetAddr())
	require.NoError(t, svc.Close())
	wg.Wait()

	{
		err := PingConn(svc.GetAddr(), 1, time.Microsecond, nil)
		eNet := err.(*net.OpError)
		require.Equal(t, eNet.Op, "dial")
		require.True(t, eNet.Timeout())
		require.Equal(t, address, eNet.Addr.String())
	}
}

func TestGRPCCloseBeforeRun(t *testing.T) {

	h, p := tempAddress(t)
	address := net.JoinHostPort(h, p)

	svc := NewGRPC(grpc.WriteBufferSize(100))
	require.NoError(t, svc.Close())

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		require.Equal(t, http.ErrServerClosed, svc.ListenAndServeAddr(nil, address))

		// test: safe close
		require.NoError(t, svc.Close())
	}()

	clientOptions := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Second)}

	require.Equal(t, context.DeadlineExceeded, PingGRPC(address, 2, clientOptions...))

	require.Equal(t, address, svc.GetAddr())
	require.NoError(t, svc.Close())
	wg.Wait()

	require.Equal(t,
		context.DeadlineExceeded,
		PingGRPC(address, 1, clientOptions...))
}

func TestGRPCWithTLS(t *testing.T) {

	l, err := logger.New()
	require.NoError(t, err)

	tlsCert, ca := newTLSCert(t)

	h, p := tempAddress(t)
	address := net.JoinHostPort(h, p)

	{
		// create server
		creds := credentials.NewTLS(&tls.Config{
			ServerName:   "127.0.0.1",
			Certificates: []tls.Certificate{*tlsCert},
		})

		svc := NewGRPC(grpc.Creds(creds))
		svc.RegisterService(func(svr *grpc.Server) {
			test.RegisterCheckerServer(svr, test.NewCheckerImpl())
		})

		var wgClose sync.WaitGroup
		wgClose.Add(1)
		go func() {
			defer wgClose.Done()

			require.NoError(t, svc.ListenAndServeAddr(l, address))
		}()

		defer func() {
			require.NoError(t, svc.Close())
			wgClose.Wait()
		}()
	}

	{
		// test: client ok
		caPool := x509.NewCertPool()
		caPool.AddCert(ca)

		clientOptions := []grpc.DialOption{
			grpc.WithBlock(),
			grpc.WithTimeout(time.Second),
			grpc.WithTransportCredentials(
				credentials.NewClientTLSFromCert(caPool, "127.0.0.1")),
		}

		require.NoError(t, PingGRPC(address, 2, clientOptions...))

		clientConn, err := grpc.Dial(address, clientOptions...)
		require.NoError(t, err)

		client := test.NewCheckerClient(clientConn)
		answer, err := client.Ping(context.Background(), &types.Empty{})
		require.NoError(t, err)
		require.Equal(t, &types.Empty{}, answer)
	}

	{
		// test: client error
		clientOptions := []grpc.DialOption{
			grpc.WithBlock(),
			grpc.WithInsecure(),
			grpc.WithTimeout(time.Second),
		}

		clientConn, err := grpc.Dial(address, clientOptions...)
		require.Equal(t, context.DeadlineExceeded, err)
		require.Nil(t, clientConn)

		require.Equal(t, context.DeadlineExceeded, PingGRPC(address, 2, clientOptions...))
	}
}

func TestHTTPWithTLS(t *testing.T) {

	l, err := logger.New()
	require.NoError(t, err)

	tlsCert, ca := newTLSCert(t)

	caPool := x509.NewCertPool()
	caPool.AddCert(ca)

	h, p := tempAddress(t)
	address := net.JoinHostPort(h, p)

	// create server
	svr := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		TLSConfig: &tls.Config{
			Certificates:             []tls.Certificate{*tlsCert},
			ServerName:               "localhost",
			RootCAs:                  caPool,
			PreferServerCipherSuites: true,
			MinVersion:               tls.VersionSSL30,
			MaxVersion:               tls.VersionTLS13,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
			},
		},
	}

	svc := NewHTTPWithServer(svr, time.Second)

	var wgClose sync.WaitGroup
	go func() {
		defer wgClose.Done()

		require.Equal(t, http.ErrServerClosed, svc.ListenAndServeAddr(l, address))
	}()

	defer func() {
		wgClose.Add(1)

		require.NoError(t, svc.Close())
		wgClose.Wait()
	}()

	for _, testInfo := range []struct {
		Name string
		Func func(*testing.T)
	}{
		{
			Name: "ping",
			Func: func(*testing.T) {
				testHTTPTLSClientPing(t, caPool, tlsCert, address)
			},
		},
		{
			Name: "ok",
			Func: func(*testing.T) {
				testHTTPTLSClientOk(t, caPool, tlsCert, address)
			},
		},
		{
			Name: "error",
			Func: func(*testing.T) {
				testHTTPTLSClientError(t, caPool, tlsCert, address)
			},
		},
	} {
		if !t.Run(testInfo.Name, testInfo.Func) {
			return
		}

		require.Equal(t, address, svc.GetAddr())
	}
}

func testHTTPTLSClientPing(t *testing.T, caPool *x509.CertPool, tlsCert *tls.Certificate, address string) {

	require.NoError(t,
		PingConn(address, 2, time.Second, nil))

	require.NoError(t,
		PingConn(address, 2, time.Second, &tls.Config{
			ServerName: "127.0.0.1",
			RootCAs:    caPool,
		}))

	require.EqualError(t,
		PingConn(address, 2, time.Second, &tls.Config{
			ServerName:   "127.0.0.1",
			RootCAs:      caPool,
			Certificates: []tls.Certificate{*tlsCert},
			CipherSuites: []uint16{tls.TLS_CHACHA20_POLY1305_SHA256},
			MaxVersion:   tls.VersionTLS10,
		}),
		"remote error: tls: handshake failure")
}

func testHTTPTLSClientError(t *testing.T, caPool *x509.CertPool, tlsCert *tls.Certificate, address string) {

	variants := map[*tls.Config]string{
		// invalid cipher suites
		&tls.Config{
			RootCAs:      caPool,
			ServerName:   "127.0.0.1",
			Certificates: []tls.Certificate{*tlsCert},
			CipherSuites: []uint16{tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA},
			MaxVersion:   tls.VersionTLS10,
		}: "Get %s: remote error: tls: handshake failure",
		// invalid server name
		&tls.Config{
			ServerName:   "localhost",
			RootCAs:      caPool,
			Certificates: []tls.Certificate{*tlsCert},
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
			},
		}: `Get "%s": x509: certificate is not valid for any names, but wanted to match localhost`,
		// invalid root CAs
		&tls.Config{
			ServerName:   "127.0.0.1",
			Certificates: []tls.Certificate{*tlsCert},
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
			},
		}: "Get %s: x509: certificate signed by unknown authority",
	}

	for tlsConf, e := range variants {
		client := http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConf,
			},
		}
		u := (&url.URL{Scheme: "https", Host: address}).String()

		resp, err := client.Get(u)
		require.Nil(t, resp, e)
		require.Contains(t,
			[]string{
				fmt.Sprintf(e, u),                      // before go1.14
				fmt.Sprintf(e, fmt.Sprintf(`"%s"`, u)), // go1.14
			},
			err.Error(),
			"error: "+e)
	}
}

func testHTTPTLSClientOk(t *testing.T, caPool *x509.CertPool, tlsCert *tls.Certificate, address string) {

	variants := []*tls.Config{
		// full
		&tls.Config{
			RootCAs:      caPool,
			ServerName:   "127.0.0.1",
			Certificates: []tls.Certificate{*tlsCert},
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
			},
		},
		// without cipher suites
		&tls.Config{
			RootCAs:      caPool,
			ServerName:   "127.0.0.1",
			Certificates: []tls.Certificate{*tlsCert},
		},
		// without tls certificate and cipher suites
		&tls.Config{
			RootCAs:    caPool,
			ServerName: "127.0.0.1",
		},
		// without tls certificate
		&tls.Config{
			RootCAs:    caPool,
			ServerName: "127.0.0.1",
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
			},
		},
	}

	for _, tlsConf := range variants {
		client := http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConf,
			},
		}
		u := (&url.URL{Scheme: "https", Host: address}).String()

		resp, err := client.Get(u)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode, getHTTPResponseBody(t, resp))
	}
}

func tempAddress(t *testing.T) (host, port string) {

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()

	host, port, err = net.SplitHostPort(l.Addr().String())
	require.NoError(t, err)
	return
}

func newTLSCert(t *testing.T) (*tls.Certificate, *x509.Certificate) {

	der, key, err := cert.NewTestCert(1024, func(ca *x509.Certificate) {
		ca.NotBefore = time.Now()
		ca.NotAfter = time.Now().AddDate(0, 0, 1)
		ca.IPAddresses = []net.IP{net.IPv4(127, 0, 0, 1)}
	}, cert.NewAttrs("localhost", "email@mydomain.com", []string{"CA"}, []string{"localhost"})...)
	require.NoError(t, err)

	tlsCert, x509Cert, err := cert.NewServerClientCerts(der, key, "123456")
	require.NoError(t, err)

	return tlsCert, x509Cert
}

func getHTTPResponseBody(t *testing.T, resp *http.Response) string {

	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	data, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	return string(data)
}
