package certs

import (
	"crypto/ecdsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/cfg"

	"github.com/zalgonoise/x/authz/internal/keygen"
)

func TestVerify(t *testing.T) {
	t.Run("AuthzOnCA", func(t *testing.T) {
		caPriv, _ := getKeys(t, "ca")
		_, authzPub := getKeys(t, "authz")

		caCert := newCACert(t, caPriv)

		authzCert := newAuthzCert(t, "authz.authz-root.test", caPriv, caCert, authzPub)

		roots := x509.NewCertPool()
		roots.AddCert(caCert)

		require.NoError(t, Verify(authzCert, caCert, nil))
	})

	t.Run("NestedOnce", func(t *testing.T) {
		caPriv, _ := getKeys(t, "ca")
		authzRootPriv, authzRootPub := getKeys(t, "authz")
		_, authzSvcPub := getKeys(t, "svc")

		caCert := newCACert(t, caPriv)

		authzSignedCert := newAuthzCert(t, "authz.authz-root.test", caPriv, caCert, authzRootPub)
		authzCert, err := Decode(authzSignedCert)
		require.NoError(t, err)

		inter := x509.NewCertPool()
		inter.AddCert(authzCert)

		signedCert := newAuthzCert(t, "authz.authz-service.test", authzRootPriv, authzCert, authzSvcPub)

		require.NoError(t, Verify(signedCert, caCert, inter))
	})

	t.Run("NestedThreeTimes", func(t *testing.T) {
		caPriv, _ := getKeys(t, "ca")
		authzRootPriv, authzRootPub := getKeys(t, "authz")
		authzSvcPriv, authzSvcPub := getKeys(t, "svc")
		authzSubSvcPriv, authzSubSvcPub := getKeys(t, "subsvc")
		_, authzSubSvc2Pub := getKeys(t, "subsvc2")

		caCert := newCACert(t, caPriv)

		authzSignedCert := newAuthzCert(t, "authz.authz-root.test", caPriv, caCert, authzRootPub)
		authzCert, err := Decode(authzSignedCert)
		require.NoError(t, err)

		inter1 := x509.NewCertPool()
		inter1.AddCert(authzCert)

		svcSignedCert := newAuthzCert(t, "authz.authz-service.test", authzRootPriv, authzCert, authzSvcPub)

		require.NoError(t, Verify(svcSignedCert, caCert, inter1))

		svcCert, err := Decode(svcSignedCert)
		require.NoError(t, err)

		inter2 := x509.NewCertPool()
		inter2.AddCert(authzCert)
		inter2.AddCert(svcCert)

		subSignedCert := newAuthzCert(t, "authz.authz-service-sub.test", authzSvcPriv, svcCert, authzSubSvcPub)

		require.NoError(t, Verify(subSignedCert, caCert, inter2))

		svc2Cert, err := Decode(subSignedCert)
		require.NoError(t, err)

		inter3 := x509.NewCertPool()
		inter3.AddCert(authzCert)
		inter3.AddCert(svcCert)
		inter3.AddCert(svc2Cert)

		signedCert := newAuthzCert(t, "authz.authz-service-sub-2.test", authzSubSvcPriv, svc2Cert, authzSubSvc2Pub)

		require.NoError(t, Verify(signedCert, caCert, inter3))
	})
}

func getKeys(t *testing.T, target string) (*ecdsa.PrivateKey, *ecdsa.PublicKey) {
	privPEM, err := os.ReadFile(fmt.Sprintf("./testdata/%s.testkey_private.pem", target))
	require.NoError(t, err)

	priv, err := keygen.DecodePrivate(privPEM)
	require.NoError(t, err)

	pubPEM, err := os.ReadFile(fmt.Sprintf("./testdata/%s.testkey_public.pem", target))
	require.NoError(t, err)

	pub, err := keygen.DecodePublic(pubPEM)
	require.NoError(t, err)

	return priv, pub
}

func newCACert(t *testing.T, key *ecdsa.PrivateKey) *x509.Certificate {
	tmpl := cfg.Set(DefaultTemplate(),
		WithName(pkix.Name{CommonName: "authz.ca.test"}),
		WithDurMonth(24),
		WithPrivateKey(key),
	)

	cert, err := NewCACertificate(tmpl)
	require.NoError(t, err)

	ca, err := Decode(cert)
	require.NoError(t, err)

	return ca
}

func newAuthzCert(t *testing.T, name string, caPrivKey *ecdsa.PrivateKey, caCert *x509.Certificate, authzPubKey *ecdsa.PublicKey) []byte {
	signingReq, err := NewCertFromCSR(caCert.Version, 24,
		caCert.Subject,
		ToCSR(name, authzPubKey, nil),
	)
	require.NoError(t, err)

	signedCert, err := Encode(signingReq, caCert, authzPubKey, caPrivKey)
	require.NoError(t, err)

	return signedCert
}
