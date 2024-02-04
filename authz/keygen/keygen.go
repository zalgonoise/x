package keygen

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"hash"
)

var (
	ErrInvalidPEM       = errors.New("invalid PEM key bytes")
	ErrInvalidSignature = errors.New("invalid signature")
)

func New(curve elliptic.Curve) (*ecdsa.PrivateKey, error) {
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
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

type Signer struct {
	privKey *ecdsa.PrivateKey
	h       hash.Hash
}

func NewSigner(privKey *ecdsa.PrivateKey, h hash.Hash) Signer {
	return Signer{privKey, h}
}

func (s Signer) Sign(data []byte) (signature []byte, err error) {
	h, err := s.Hash(data)
	if err != nil {
		return nil, err
	}

	return ecdsa.SignASN1(rand.Reader, s.privKey, h)
}

func (s Signer) Hash(data []byte) (hash []byte, err error) {
	defer s.h.Reset()

	_, err = s.h.Write(data)
	if err != nil {
		return nil, err
	}

	h := s.h.Sum(nil)

	return h, nil
}

type Verifier struct {
	pubKey *ecdsa.PublicKey
	h      hash.Hash
}

func NewVerifier(pubKey *ecdsa.PublicKey, h hash.Hash) Verifier {
	return Verifier{pubKey, h}
}

func (v Verifier) Verify(data []byte, signature []byte) error {
	h, err := v.Hash(data)
	if err != nil {
		return err
	}

	if !ecdsa.VerifyASN1(v.pubKey, h, signature) {
		return ErrInvalidSignature
	}

	return nil
}

func (v Verifier) Hash(data []byte) (hash []byte, err error) {
	defer v.h.Reset()

	_, err = v.h.Write(data)
	if err != nil {
		return nil, err
	}

	h := v.h.Sum(nil)

	return h, nil
}
