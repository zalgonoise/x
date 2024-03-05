package keygen

import (
	"crypto/ecdsa"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/zalgonoise/cfg"
)

func NewToken(privateKey *ecdsa.PrivateKey, issuer string, expiry time.Time, opts ...cfg.Option[JWT]) ([]byte, error) {
	if privateKey == nil {
		return nil, ErrNilPrivateKey
	}

	options := cfg.New(opts...)

	b := jwt.NewBuilder().Issuer(issuer).Expiration(expiry)

	if options.Subject != "" {
		b.Subject(options.Subject)
	}

	if options.claimName != "" {
		b.Claim(options.claimName, options.Claim)
	}

	if len(options.Audience) > 0 {
		b.Audience(options.Audience)
	}

	if options.ID != "" {
		b.JwtID(options.ID)
	}

	var zero time.Time
	if options.NotBefore != zero && !options.NotBefore.IsZero() {
		b.NotBefore(options.NotBefore)
	}

	token, err := b.Build()
	if err != nil {
		return nil, err
	}

	signedToken, err := jwt.Sign(token, jwt.WithKey(jwa.ES512, privateKey))
	if err != nil {
		return nil, err
	}

	return signedToken, nil
}

func marshal(m map[string]any) Claim {
	claim := Claim{}

	if v, ok := m["service"]; ok {
		if value, ok := v.(string); ok {
			claim.Service = value
		}
	}

	if v, ok := m["authz_service"]; ok {
		if value, ok := v.(string); ok {
			claim.Authz = value
		}
	}

	return claim
}

func ParseToken(token []byte, publicKey *ecdsa.PublicKey) (JWT, error) {
	opt := jwt.WithVerify(false)

	if publicKey != nil {
		opt = jwt.WithKey(jwa.ES512, publicKey)
	}

	t, err := jwt.Parse(token, opt)
	if err != nil {
		return JWT{}, err
	}

	c, ok := t.Get(authzClaim)
	if !ok {
		return JWT{}, ErrInvalidClaim
	}

	claimMap, ok := c.(map[string]any)
	if !ok {
		return JWT{}, ErrInvalidClaim
	}

	return JWT{
		Issuer:    t.Issuer(),
		Subject:   t.Subject(),
		claimName: authzClaim,
		Claim:     marshal(claimMap),
		Audience:  t.Audience(),
		ID:        t.JwtID(),
		NotBefore: t.NotBefore(),
		Expiry:    t.Expiration(),
	}, nil
}
