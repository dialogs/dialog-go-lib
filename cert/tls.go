package cert

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	pkgerr "github.com/pkg/errors"
)

func NewTlsConfig(certFile, keyFile, caFile, serverName string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, pkgerr.Wrap(err, "load x509 key pair failed")
	}

	ca, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, pkgerr.Wrap(err, "load ca-cert failed")
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(ca)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		ServerName:   serverName,
	}
	tlsConfig.BuildNameToCertificate()
	return tlsConfig, nil
}
