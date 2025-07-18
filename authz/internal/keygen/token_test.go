package keygen

import (
	_ "embed"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/cfg"
)

//go:embed testdata/testkey_private.pem
var privateKey []byte

//go:embed testdata/testkey_public.pem
var publicKey []byte

func TestNewToken(t *testing.T) {
	priv, err := DecodePrivate(privateKey)
	require.NoError(t, err)

	pub, err := DecodePublic(publicKey)
	require.NoError(t, err)

	exp := time.Date(2030, 12, 31, 23, 59, 59, 0, time.UTC)

	for _, testcase := range []struct {
		name   string
		issuer string
		exp    time.Time
		opts   []cfg.Option[JWT]
	}{
		{
			name:   "simple",
			issuer: "authz_test",
			exp:    exp,
			opts: []cfg.Option[JWT]{
				WithClaim(Claim{
					Service: "test",
					Authz:   "authz.test",
				}),
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			token, err := NewToken(priv, testcase.issuer, testcase.exp, testcase.opts...)
			require.NoError(t, err)

			t.Logf("[%s]\t token: %s\n\n", testcase.name, string(token))

			jwtData, err := ParseToken(token, pub)
			require.NoError(t, err)

			t.Logf("[%s]\t token data: %+v\n\n", testcase.name, jwtData)
		})
	}
}
