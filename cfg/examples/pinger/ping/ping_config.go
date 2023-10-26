package ping

import (
	"fmt"
	"net/url"
	"time"

	"github.com/zalgonoise/x/cfg"
	"github.com/zalgonoise/x/errs"
)

const defaultTimeout = 15 * time.Second

var (
	defaultConfig = Config{
		timeout: defaultTimeout,
	}
)

const (
	errDomain = errs.Domain("x/cfg/examples/pinger/ping")

	ErrEmpty   = errs.Kind("empty")
	ErrInvalid = errs.Kind("invalid")

	ErrURL = errs.Entity("URL")
)

var (
	ErrEmptyURL   = errs.WithDomain(errDomain, ErrEmpty, ErrURL)
	ErrInvalidURL = errs.WithDomain(errDomain, ErrInvalid, ErrURL)
)

type Config struct {
	url     string
	timeout time.Duration
}

func WithURL(url string) cfg.Option[Config] {
	if url == "" {
		return cfg.NoOp[Config]{}
	}

	// register an option by declaring the returned function as a ConfigFunc type
	return cfg.ConfigFunc[Config](func(config Config) Config {
		config.url = url

		return config
	})
}

// validation can be handled separately, via its own module, or used as part of the
// NewChecker constructor
func validateURL(config Config) error {
	if config.url == "" {
		return ErrEmptyURL
	}

	if _, err := url.Parse(config.url); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidURL, err)
	}

	return nil
}

func WithTimeout(dur time.Duration) cfg.Option[Config] {
	if dur <= 0 {
		return cfg.NoOp[Config]{}
	}

	// register an option via the cfg.Register function
	return cfg.Register(func(config Config) Config {
		config.timeout = dur

		return config
	})
}
