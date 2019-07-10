package cert

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"math/big"
	"time"

	pkgerr "github.com/pkg/errors"
	"golang.org/x/crypto/pkcs12"
	gopkcs12 "software.sslmate.com/src/go-pkcs12"
)

// Certificate attributes oid:
// source:
// - https://ldap.com/ldap-oid-reference-guide/
// - http://gmssl.org/docs/oid.html
var (
	oidDN    = asn1.ObjectIdentifier{0, 9, 2342, 19200300, 100, 1, 25} // example: ru
	oidOU    = asn1.ObjectIdentifier{2, 5, 4, 11}                      // example: CA, CA_Users
	oidCN    = asn1.ObjectIdentifier{2, 5, 4, 3}                       // example: username
	oidEmail = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 1}       // example: email@domain.ru
)

// NewAttrs returns certificate attributes list
func NewAttrs(cn, email string, ou, dn []string) []pkix.AttributeTypeAndValue {

	list := []pkix.AttributeTypeAndValue{}
	if cn != "" {
		list = append(list, pkix.AttributeTypeAndValue{
			Type:  oidCN,
			Value: cn,
		})
	}

	if email != "" {
		list = append(list, pkix.AttributeTypeAndValue{
			Type:  oidEmail,
			Value: email,
		})
	}

	if len(ou) > 0 {
		// example: CA, CA_Users
		for _, item := range ou {
			list = append(list, pkix.AttributeTypeAndValue{
				Type:  oidOU,
				Value: item,
			})
		}
	}

	if len(dn) > 0 {
		// example: ru, mydomain
		for _, item := range dn {
			list = append(list, pkix.AttributeTypeAndValue{
				Type:  oidDN,
				Value: item,
			})
		}
	}

	return list
}

// NewX509 returns x509 certificate
func NewX509(number *big.Int, start, end time.Time, dnsNames []string, attrs []pkix.AttributeTypeAndValue) *x509.Certificate {

	return &x509.Certificate{
		SerialNumber: number,
		Issuer: pkix.Name{
			ExtraNames: attrs,
		},
		DNSNames: dnsNames,
		Subject: pkix.Name{
			ExtraNames: attrs,
		},
		NotBefore: start,
		NotAfter:  end,
		IsCA:      false,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth,
		},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		BasicConstraintsValid: true,
	}
}

// NewRSA returns private rsa key
func NewRSA(size int) (*rsa.PrivateKey, error) {

	privateKey, err := rsa.GenerateKey(rand.Reader, size)
	if err != nil {
		return nil, pkgerr.Wrap(err, "new rsa")
	}

	return privateKey, nil
}

// X509ToDerBytes converts x509 to der
func X509ToDerBytes(template, parent *x509.Certificate, privateKey *rsa.PrivateKey) ([]byte, error) {

	derBytes, err := x509.CreateCertificate(
		rand.Reader,
		template, parent,
		&privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, pkgerr.Wrap(err, "x509 to der format")
	}

	return derBytes, nil
}

// P12ToCert converts p12(pfx) certificate to x509
func P12ToCert(p12 []byte, pass string) (interface{}, *x509.Certificate, error) {

	privateKey, cert, err := pkcs12.Decode(p12, pass)
	if err != nil {
		return nil, nil, pkgerr.Wrap(err, "decode pfx")
	}

	return privateKey, cert, nil
}

// X509ToP12 converts x509 certificate to p12(pfx) format
func X509ToP12(der []byte, privateKey *rsa.PrivateKey, pass string) ([]byte, error) {

	domainCert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, pkgerr.Wrap(err, "x509 to p12: parse certificate")
	}

	pfxData, err := gopkcs12.Encode(rand.Reader, privateKey, domainCert, nil, pass)
	if err != nil {
		return nil, pkgerr.Wrap(err, "x509 to p12: encode to pfx")
	}

	return pfxData, nil
}

// DerToPem returns x509 certificate in pem format
func DerToPem(der []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: der,
	})
}

// PemToX509 returns x509 certificate from pem
func PemToX509(certPEM []byte) (*x509.Certificate, error) {

	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return nil, errors.New("pem to x509: decode pem")
	}

	c, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, pkgerr.Wrap(err, "pem to x509: parse der")
	}

	return c, nil
}

// RsaToPem converts private rsa key to pem format
func RsaToPem(key *rsa.PrivateKey) []byte {

	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
}

// PemToRsa converts pem to rsa private key
func PemToRsa(pemData []byte) (*rsa.PrivateKey, error) {

	pemBlock, _ := pem.Decode(pemData)
	if pemBlock == nil {
		return nil, errors.New("pem to rsa: decode pem")
	}

	rsaKey, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
	if err != nil {
		return nil, pkgerr.Wrap(err, "pem to rsa")
	}

	return rsaKey, nil
}

// P12ToTLS converts p12(pfx) to TLS certificate
func P12ToTLS(p12 []byte, password string) (*tls.Certificate, error) {

	blocks, err := pkcs12.ToPEM(p12, password)
	if err != nil {
		return nil, pkgerr.Wrap(err, "p12 to TLS: p12 to pem")
	}

	pemData := bytes.NewBuffer(nil)
	for _, b := range blocks {
		pemData.Write(pem.EncodeToMemory(b))
	}

	tlsKey, err := tls.X509KeyPair(pemData.Bytes(), pemData.Bytes())
	if err != nil {
		return nil, pkgerr.Wrap(err, "p12 to TLS: failed to create TLS certificate")
	}

	return &tlsKey, nil
}
