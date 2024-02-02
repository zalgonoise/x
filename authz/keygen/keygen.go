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
	Hash hash.Hash
}

func NewSigner(h hash.Hash) Signer {
	return Signer{h}
}

func (s Signer) Verify(data []byte, pubKey *ecdsa.PublicKey, signature []byte) error {
	defer s.Hash.Reset()

	_, err := s.Hash.Write(data)
	if err != nil {
		return err
	}

	signatureHash := s.Hash.Sum(nil)

	if !ecdsa.VerifyASN1(pubKey, signatureHash, signature) {
		return ErrInvalidSignature
	}

	return nil
}

func (s Signer) Sign(data []byte, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	defer s.Hash.Reset()

	_, err := s.Hash.Write(data)
	if err != nil {
		return nil, err
	}

	signatureHash := s.Hash.Sum(nil)

	signature, err := ecdsa.SignASN1(rand.Reader, privateKey, signatureHash)

	return signature, nil
}
