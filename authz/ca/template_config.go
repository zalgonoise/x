package ca

import (
	"crypto/ecdsa"
	"crypto/x509/pkix"
	"math/big"

	"github.com/zalgonoise/cfg"
)

func defaultTemplate() Template {
	return Template{
		DurMonth:  defaultDurMonths,
		SerialExp: defaultExp,
		SerialSub: defaultSub,
	}
}

func WithName(name pkix.Name) cfg.Option[Template] {
	if name.CommonName == "" {
		return cfg.NoOp[Template]{}
	}

	return cfg.Register[Template](func(template Template) Template {
		template.Name = name

		return template
	})
}

func WithDurMonth(durMonth int) cfg.Option[Template] {
	if durMonth == 0 {
		durMonth = defaultDurMonths
	}

	return cfg.Register[Template](func(template Template) Template {
		template.DurMonth = durMonth

		return template
	})
}

func WithPrivateKey(key *ecdsa.PrivateKey) cfg.Option[Template] {
	if key == nil {
		return cfg.NoOp[Template]{}
	}

	return cfg.Register[Template](func(template Template) Template {
		template.PrivateKey = key

		return template
	})
}

func WithSerial(i *big.Int) cfg.Option[Template] {
	if i == nil {
		return cfg.NoOp[Template]{}
	}

	return cfg.Register[Template](func(template Template) Template {
		template.Serial = i

		return template
	})
}

func WithNewSerial(exp, sub int64) cfg.Option[Template] {
	if exp == 0 {
		exp = defaultExp
	}

	if sub == 0 {
		sub = defaultSub
	}

	return cfg.Register[Template](func(template Template) Template {
		template.SerialExp = exp
		template.SerialSub = sub

		return template
	})
}
