package httpaudio

import (
	"strings"
	"time"

	"github.com/zalgonoise/x/audio/errs"
	"github.com/zalgonoise/x/audio/validation"
	"github.com/zalgonoise/x/cfg"
)

const (
	defaultConnTimeout = 30 * time.Second
	protoHTTP          = "http://"
	protoHTTPS         = "https://"

	consumerDomain = errs.Domain("audio/sdk/audio/consumers/httpaudio")

	ErrEmpty   = errs.Kind("empty")
	ErrInvalid = errs.Kind("invalid")

	ErrAddress    = errs.Entity("address")
	ErrTimeoutDur = errs.Entity("timeout duration")
	ErrProtocol   = errs.Entity("protocol")
)

var (
	ErrEmptyAddress    = errs.New(consumerDomain, ErrEmpty, ErrAddress)
	ErrEmptyTimeoutDur = errs.New(consumerDomain, ErrEmpty, ErrTimeoutDur)
	ErrInvalidProtocol = errs.New(consumerDomain, ErrInvalid, ErrProtocol)

	configValidator = validation.Register[HTTPConfig](validateTarget, validateDuration)
	defaultConfig   = HTTPConfig{
		timeout: defaultConnTimeout,
	}
)

// HTTPConfig defines a data structure for configurations and options related to a HTTP audio.Consumer.
type HTTPConfig struct {
	target  string
	timeout time.Duration
}

// WithTimeout sets a general timeout for the HTTP connection.
func WithTimeout(dur time.Duration) cfg.Option[HTTPConfig] {
	return cfg.Register(func(config HTTPConfig) HTTPConfig {
		config.timeout = dur

		return config
	})
}

func validateDuration(config HTTPConfig) error {
	if config.timeout == 0 {
		return ErrEmptyTimeoutDur
	}

	return nil
}

// WithTarget defines the HTTP URL of the audio source.
func WithTarget(target string) cfg.Option[HTTPConfig] {
	return cfg.Register(func(config HTTPConfig) HTTPConfig {
		config.target = target

		return config
	})
}

func validateTarget(config HTTPConfig) error {
	if config.target == "" {
		return ErrEmptyAddress
	}

	if !strings.HasPrefix(config.target, protoHTTP) ||
		!strings.HasPrefix(config.target, protoHTTPS) {
		return ErrInvalidProtocol
	}

	return nil
}

// Validate verifies if the input HTTPConfig contains missing or invalid fields
func Validate(config HTTPConfig) error {
	return configValidator.Validate(config)
}
