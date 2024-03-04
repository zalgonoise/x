package keygen

import (
	"time"

	"github.com/zalgonoise/cfg"
)

const authzClaim = "authz"

type JWT struct {
	Issuer    string
	Subject   string
	claimName string
	Claim     Claim
	Audience  []string
	ID        string
	Expiry    time.Time
	NotBefore time.Time
}

type Claim struct {
	Service string `json:"service"`
	Authz   string `json:"authz_service"`
}

func WithSubject(subject string) cfg.Option[JWT] {
	if subject == "" {
		return cfg.NoOp[JWT]{}
	}

	return cfg.Register[JWT](func(jwt JWT) JWT {
		jwt.Subject = subject

		return jwt
	})
}

func WithClaim(claim Claim) cfg.Option[JWT] {
	if claim.Service == "" {
		return cfg.NoOp[JWT]{}
	}

	return cfg.Register[JWT](func(jwt JWT) JWT {
		jwt.claimName = authzClaim
		jwt.Claim = claim

		return jwt
	})
}

func WithAudience(audience []string) cfg.Option[JWT] {
	if len(audience) == 0 {
		return cfg.NoOp[JWT]{}
	}

	return cfg.Register[JWT](func(jwt JWT) JWT {
		jwt.Audience = audience

		return jwt
	})
}

func WithID(id string) cfg.Option[JWT] {
	if id == "" {
		return cfg.NoOp[JWT]{}
	}

	return cfg.Register[JWT](func(jwt JWT) JWT {
		jwt.ID = id

		return jwt
	})
}

func WithNotBefore(notBefore time.Time) cfg.Option[JWT] {
	var zero time.Time

	if notBefore == zero || notBefore.IsZero() {
		return cfg.NoOp[JWT]{}
	}

	return cfg.Register[JWT](func(jwt JWT) JWT {
		jwt.NotBefore = notBefore

		return jwt
	})
}
