package keygen

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha512"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/zalgonoise/x/errs"
)

const typeCertificate = "CERTIFICATE"

const (
	errDomain = errs.Domain("x/authz/keygen")

	ErrNil     = errs.Kind("nil")
	ErrInvalid = errs.Kind("invalid")

	ErrPrivateKey = errs.Entity("private key")
	ErrPEM        = errs.Entity("PEM key bytes")
	ErrSignature  = errs.Entity("signature")
)

var (
	ErrInvalidPEM       = errs.WithDomain(errDomain, ErrInvalid, ErrPEM)
	ErrInvalidSignature = errs.WithDomain(errDomain, ErrInvalid, ErrSignature)
	ErrNilPrivateKey    = errs.WithDomain(errDomain, ErrNil, ErrPrivateKey)
)

func New() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
}

func EncodePrivate(privateKey *ecdsa.PrivateKey) (key []byte, err error) {
	encoded, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	return pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: encoded}), nil
}

func EncodePublic(publicKey *ecdsa.PublicKey) (key []byte, err error) {
	encoded, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return
	}

	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: encoded}), nil
}

func DecodePrivate(pemPrivate []byte) (privateKey *ecdsa.PrivateKey, err error) {
	pemBlock, _ := pem.Decode(pemPrivate)
	if pemBlock == nil {
		return nil, ErrInvalidPEM
	}

	return x509.ParseECPrivateKey(pemBlock.Bytes)
}

func DecodePublic(pemEncodedPub []byte) (*ecdsa.PublicKey, error) {
	pemBlock, _ := pem.Decode(pemEncodedPub)
	if pemBlock == nil {
		return nil, ErrInvalidPEM
	}

	publicKey, err := x509.ParsePKIXPublicKey(pemBlock.Bytes)
	if err != nil {
		return nil, err
	}

	pubKey, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("%w: public key type %T is not of type *ecdsa.PublicKey", ErrInvalidPEM, publicKey)
	}

	return pubKey, nil
}

func EncodeCertificate(template, parent *x509.Certificate, pub *ecdsa.PublicKey, priv *ecdsa.PrivateKey) ([]byte, error) {
	signedCertBytes, err := x509.CreateCertificate(rand.Reader, template, parent, pub, priv)
	if err != nil {
		return nil, err
	}

	return pem.EncodeToMemory(&pem.Block{Type: typeCertificate, Bytes: signedCertBytes}), nil
}

func DecodeCertificate(cert []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(cert)
	if block == nil {
		return nil, ErrInvalidPEM
	}

	return x509.ParseCertificate(block.Bytes)
}

type ECDSASigner struct {
	Priv *ecdsa.PrivateKey
}

func (e ECDSASigner) Sign(data []byte) (sig, hash []byte, err error) {
	sum := sha512.Sum512(data)

	sig, err = ecdsa.SignASN1(rand.Reader, e.Priv, sum[:])
	if err != nil {
		return nil, nil, err
	}

	return sig, sum[:], nil
}

type ECDSAVerifier struct {
	Pub *ecdsa.PublicKey
}

func (d ECDSAVerifier) Verify(hash, signature []byte) error {
	if !ecdsa.VerifyASN1(d.Pub, hash, signature) {
		return ErrInvalidSignature
	}

	return nil
}
