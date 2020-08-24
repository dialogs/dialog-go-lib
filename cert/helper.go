package cert

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"math/rand"
	"time"

	"github.com/pkg/errors"
)

func NewTestCert(keySize int, upgrade func(*x509.Certificate), attrs ...pkix.AttributeTypeAndValue) (der []byte, key *rsa.PrivateKey, err error) {

	ca := NewX509(big.NewInt(rand.Int63()), time.Time{}, time.Time{}, nil, attrs)
	if upgrade != nil {
		upgrade(ca)
	}

	key, err = NewRSA(keySize)
	if err != nil {
		err = errors.Wrap(err, "failed to create rsa key")
		return
	}

	der, err = X509ToDerBytes(ca, ca, key)
	if err != nil {
		err = errors.Wrap(err, "failed to convert x509 to der")
		return
	}

	return
}

func NewServerClientCerts(der []byte, key *rsa.PrivateKey, password string) (*tls.Certificate, *x509.Certificate, error) {

	client, err := DerToX509(der)
	if err != nil {
		return nil, nil, err
	}

	server, err := NewTLS(der, key, password)
	if err != nil {
		return nil, nil, err
	}

	return server, client, nil
}
