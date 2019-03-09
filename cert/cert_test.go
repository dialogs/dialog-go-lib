package cert

import (
	"crypto/tls"
	"crypto/x509"
	"math/big"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/dialogs/dialog-go-lib/service"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials"
)

func TestCert(t *testing.T) {

	start := time.Now().In(time.UTC).Truncate(time.Second)
	end := start.AddDate(0, 0, 1)

	caParent := NewX509(big.NewInt(rand.Int63()), start, end, nil,
		NewAttrs("cn1", "email1", nil, nil))

	template := NewX509(big.NewInt(rand.Int63()), start, end, nil,
		NewAttrs("cn2", "email2", []string{"CA"}, []string{"ru"}))

	privateKey, err := NewRSA(1024)
	require.NoError(t, err)

	{
		resRsa, err := PemToRsa(RsaToPem(privateKey))
		require.NoError(t, err)
		require.Equal(t, privateKey, resRsa)
	}

	derBytes, err := X509ToDerBytes(template, caParent, privateKey)
	require.NoError(t, err)

	{
		pem := DerToPem(derBytes)

		resCa, err := PemToX509(pem)
		require.NoError(t, err)

		require.Equal(t, start, resCa.NotBefore)
		require.Equal(t, end, resCa.NotAfter)
		require.Equal(t, "cn2", resCa.Subject.CommonName)
	}

	{
		p12, err := X509ToP12(derBytes, privateKey, "12345")
		require.NoError(t, err)

		key, resCa, err := P12ToCert(p12, "12345")
		require.NoError(t, err)
		require.Equal(t, privateKey, key)
		require.Equal(t, "cn2", resCa.Subject.CommonName)
		require.Equal(t, start, resCa.NotBefore)
		require.Equal(t, end, resCa.NotAfter)
	}

}

func TestGRPC(t *testing.T) {

	host := "localhost"
	attrs := NewAttrs(
		host,
		"email@"+host,
		[]string{"CA", "CA_Users"},
		nil,
	)

	start := time.Now().In(time.UTC).Truncate(time.Second)
	end := start.AddDate(0, 0, 1)

	ca := NewX509(big.NewInt(rand.Int63()), start, end, []string{host}, attrs)

	privateKey, err := NewRSA(1024)
	require.NoError(t, err)

	servDerBytes, err := X509ToDerBytes(ca, ca, privateKey)
	require.NoError(t, err)

	servCertPEMBlock := DerToPem(servDerBytes)
	certificate, err := tls.X509KeyPair(servCertPEMBlock, RsaToPem(privateKey))
	require.NoError(t, err)

	creds := credentials.NewTLS(&tls.Config{
		ServerName:   host,
		Certificates: []tls.Certificate{certificate},
	})

	grpcSvr := service.NewGRPC(grpc.Creds(creds))

	_, p := tempAddress(t)
	address := host + ":" + p

	go func() {
		require.NoError(t, grpcSvr.ListenAndServeAddr(address))
	}()
	defer grpcSvr.Close()

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(servCertPEMBlock)

	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTimeout(time.Second * 2),
		grpc.WithBalancerName(roundrobin.Name),
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(pool, "")),
	}
	require.NoError(t, service.PingGRPC(address, 2, opts...))
}

func tempAddress(t *testing.T) (host, port string) {
	t.Helper()

	l, err := net.Listen("tcp", "0.0.0.0:0")
	require.NoError(t, err)
	defer l.Close()

	host, port, err = net.SplitHostPort(l.Addr().String())
	require.NoError(t, err)
	return
}
