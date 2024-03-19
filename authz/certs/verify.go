package certs

import (
	"crypto/x509"
	"fmt"
)

func Verify(certPEM []byte, inter, root *x509.Certificate) error {
	if certPEM == nil {
		return ErrNilCertificate
	}

	certificate, err := Decode(certPEM)
	if err != nil {
		return err
	}

	if inter == nil && root == nil {
		return ErrNilCACertificate
	}

	opts := x509.VerifyOptions{
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	if root != nil {
		pool := x509.NewCertPool()
		pool.AddCert(root)

		opts.Roots = pool
	}

	if inter != nil {
		pool := x509.NewCertPool()
		pool.AddCert(inter)

		opts.Intermediates = pool
	}

	chains, err := certificate.Verify(opts)
	if err != nil {
		return err
	}

	l := len(chains)

	switch {
	case inter != nil && root == nil && l == 1,
		inter == nil && root != nil && l == 1,
		inter != nil && root != nil && l == 2:
		return nil
	default:
		return fmt.Errorf("%w: %d", ErrInvalidNumChains, l)
	}
}
