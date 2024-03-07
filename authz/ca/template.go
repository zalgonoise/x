package ca

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"
)

const (
	defaultExp       int64 = 130
	defaultSub       int64 = 1
	defaultDurMonths int   = 24

	typeCertificate = "CERTIFICATE"
)

type Template struct {
	Name       pkix.Name
	DurMonth   int
	PrivateKey *ecdsa.PrivateKey

	Serial    *big.Int
	SerialExp int64
	SerialSub int64
}

func NewCACertificate(t Template) (ca *x509.Certificate, cert *pem.Block, err error) {
	if t.Serial == nil {
		bigInt, err := newInt(2, t.SerialExp, t.SerialSub)
		if err != nil {
			return nil, nil, err
		}

		t.Serial = bigInt
	}

	ca = &x509.Certificate{
		SerialNumber:          t.Serial,
		Subject:               t.Name,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, t.DurMonth, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	data, err := x509.CreateCertificate(rand.Reader, ca, ca, &t.PrivateKey.PublicKey, t.PrivateKey)
	if err != nil {
		return nil, nil, err
	}

	return ca, &pem.Block{Type: typeCertificate, Bytes: data}, nil
}

func newInt(base, exp, sub int64) (*big.Int, error) {
	maximum := new(big.Int)
	maximum.Exp(big.NewInt(base), big.NewInt(exp), nil).Sub(maximum, big.NewInt(sub))

	return rand.Int(rand.Reader, maximum)
}

func NewCertFromCSR(version, durMonth int, csr *x509.CertificateRequest) (*x509.Certificate, error) {
	i, err := newInt(2, defaultExp, defaultSub)
	if err != nil {
		return nil, err
	}

	return &x509.Certificate{
		Version:         version,
		SerialNumber:    i,
		Subject:         csr.Subject,
		Extensions:      csr.Extensions,
		ExtraExtensions: csr.ExtraExtensions,
		DNSNames:        csr.DNSNames,
		EmailAddresses:  csr.EmailAddresses,
		IPAddresses:     csr.IPAddresses,
		URIs:            csr.URIs,
		NotBefore:       time.Now(),
		NotAfter:        time.Now().AddDate(0, durMonth, 0),
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageCodeSigning,
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageOCSPSigning,
			x509.ExtKeyUsageCodeSigning,
		},
		KeyUsage: x509.KeyUsageCertSign,
	}, nil
}
