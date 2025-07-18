package certs

import (
	"crypto/x509"
)

func Verify(certPEM []byte, root *x509.Certificate, intermediates *x509.CertPool) error {
	if certPEM == nil {
		return ErrNilCertificate
	}

	certificate, err := Decode(certPEM)
	if err != nil {
		return err
	}

	if root == nil {
		return ErrNilCACertificate
	}

	opts := x509.VerifyOptions{
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		Intermediates: intermediates,
	}

	if root != nil {
		opts.Roots = x509.NewCertPool()
		opts.Roots.AddCert(root)
	}

	_, err = certificate.Verify(opts)

	return err
}
