package service

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/dialogs/dialog-go-lib/cert"
	"github.com/dialogs/dialog-go-lib/service/test"
	"github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func init() {
	// for debug
	log.SetFlags(log.Llongfile | log.Ltime | log.Lmicroseconds)
}

func TestGRPC(t *testing.T) {

	h, p := tempAddress(t)
	address := net.JoinHostPort(h, p)

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
	address := net.JoinHostPort(h, p)

	svc := NewHTTP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), time.Second)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		require.Equal(t, http.ErrServerClosed, svc.ListenAndServeAddr(address))
	}()

	require.NoError(t, PingConn(address, 2, time.Second, nil))

	require.Equal(t, address, svc.GetAddr())
	require.NoError(t, svc.Close())
	wg.Wait()

	require.EqualError(t,
		PingConn(svc.GetAddr(), 1, time.Microsecond, nil),
		fmt.Sprintf("dial tcp %s: i/o timeout", address))
}

func TestGRPCWithTLS(t *testing.T) {

	tlsCert, ca := newTLSCert(t)

	h, p := tempAddress(t)
	address := net.JoinHostPort(h, p)

	{
		// create server
		creds := credentials.NewTLS(&tls.Config{
			ServerName:   "localhost",
			Certificates: []tls.Certificate{*tlsCert},
		})

		svc := NewGRPC(grpc.Creds(creds))
		svc.RegisterService(func(svr *grpc.Server) {
			test.RegisterCheckerServer(svr, test.NewCheckerImpl())
		})

		go func() {
			require.NoError(t, svc.ListenAndServeAddr(address))
		}()

		defer func() {
			require.NoError(t, svc.Close())
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
				credentials.NewClientTLSFromCert(caPool, "localhost")),
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

	go func() {
		require.Equal(t, http.ErrServerClosed, svc.ListenAndServeAddr(address))
	}()

	defer func() {
		require.NoError(t, svc.Close())
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
	t.Helper()

	require.NoError(t,
		PingConn(address, 2, time.Second, nil))

	require.NoError(t,
		PingConn(address, 2, time.Second, &tls.Config{
			ServerName: "localhost",
			RootCAs:    caPool,
		}))

	require.EqualError(t,
		PingConn(address, 2, time.Second, &tls.Config{
			ServerName:   "localhost",
			RootCAs:      caPool,
			CipherSuites: []uint16{},
		}),
		"remote error: tls: handshake failure")
}

func testHTTPTLSClientError(t *testing.T, caPool *x509.CertPool, tlsCert *tls.Certificate, address string) {
	t.Helper()

	variants := map[*tls.Config]string{
		// invalid cipher suites
		&tls.Config{
			RootCAs:      caPool,
			ServerName:   "localhost",
			Certificates: []tls.Certificate{*tlsCert},
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			},
		}: "Get %s: remote error: tls: handshake failure",
		// invalid server name
		&tls.Config{
			RootCAs:      caPool,
			Certificates: []tls.Certificate{*tlsCert},
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
			},
		}: "Get %s: x509: cannot validate certificate for 127.0.0.1 because it doesn't contain any IP SANs",
		// invalid root CAs
		&tls.Config{
			ServerName:   "localhost",
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
		require.EqualError(t, err, fmt.Sprintf(e, u), "error: "+e)
		require.Nil(t, resp)
	}
}

func testHTTPTLSClientOk(t *testing.T, caPool *x509.CertPool, tlsCert *tls.Certificate, address string) {
	t.Helper()

	variants := []*tls.Config{
		// full
		&tls.Config{
			RootCAs:      caPool,
			ServerName:   "localhost",
			Certificates: []tls.Certificate{*tlsCert},
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
			},
		},
		// without cipher suites
		&tls.Config{
			RootCAs:      caPool,
			ServerName:   "localhost",
			Certificates: []tls.Certificate{*tlsCert},
		},
		// without tls certificate and cipher suites
		&tls.Config{
			RootCAs:    caPool,
			ServerName: "localhost",
		},
		// without tls certificate
		&tls.Config{
			RootCAs:    caPool,
			ServerName: "localhost",
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
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()

	host, port, err = net.SplitHostPort(l.Addr().String())
	require.NoError(t, err)
	return
}

func newTLSCert(t *testing.T) (*tls.Certificate, *x509.Certificate) {
	t.Helper()

	ca := cert.NewX509(
		big.NewInt(rand.Int63()),
		time.Now(), time.Now().AddDate(0, 0, 1),
		nil,
		cert.NewAttrs(
			"localhost",
			"email@mydomain.com",
			[]string{"CA"},
			[]string{"localhost"}))

	privateKey, err := cert.NewRSA(1024)
	require.NoError(t, err)

	derBytes, err := cert.X509ToDerBytes(ca, ca, privateKey)
	require.NoError(t, err)

	const CertPassword = "123456"

	p12, err := cert.X509ToP12(derBytes, privateKey, CertPassword)
	require.NoError(t, err)

	tlsCert, err := cert.P12ToTLS(p12, CertPassword)
	require.NoError(t, err)

	pem := cert.DerToPem(derBytes)
	clientCert, err := cert.PemToX509(pem)
	require.NoError(t, err)

	return tlsCert, clientCert
}

func getHTTPResponseBody(t *testing.T, resp *http.Response) string {
	t.Helper()

	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	data, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	return string(data)
}
