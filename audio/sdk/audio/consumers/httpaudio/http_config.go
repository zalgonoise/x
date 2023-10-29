package httpaudio

import (
	"strings"
	"time"

	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/valigator"
	"github.com/zalgonoise/x/errs"
)

const (
	defaultConnTimeout = 30 * time.Second
	protoHTTP          = "http://"
	protoHTTPS         = "https://"

	consumerDomain = errs.Domain("audio/sdk/audio/consumers/httpaudio")

	ErrInvalid = errs.Kind("invalid")

	ErrProtocol = errs.Entity("protocol")
)

var (
	ErrInvalidProtocol = errs.WithDomain(consumerDomain, ErrInvalid, ErrProtocol)

	configValidator = valigator.New(validateTarget)
	defaultConfig   = Config{
		timeout: defaultConnTimeout,
	}
)

// Config defines a data structure for configurations and options related to a HTTP audio.Consumer.
type Config struct {
	target  string
	timeout time.Duration
}

// WithTimeout sets a general timeout for the HTTP connection.
func WithTimeout(dur time.Duration) cfg.Option[Config] {
	if dur == 0 {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register(func(config Config) Config {
		config.timeout = dur

		return config
	})
}

// WithTarget defines the HTTP URL of the audio source.
func WithTarget(target string) cfg.Option[Config] {
	if target == "" {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register(func(config Config) Config {
		config.target = target

		return config
	})
}

func validateTarget(config Config) error {
	if !strings.HasPrefix(config.target, protoHTTP) &&
		!strings.HasPrefix(config.target, protoHTTPS) {
		return ErrInvalidProtocol
	}

	return nil
}

// Validate verifies if the input Config contains missing or invalid fields
func Validate(config Config) error {
	return configValidator.Validate(config)
}
