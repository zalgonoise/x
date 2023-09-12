package webhook

import (
	"time"

	"github.com/zalgonoise/x/cfg"
	"github.com/zalgonoise/x/errs"
)

const (
	errDomain = errs.Domain("x/discord/webhook")

	ErrEmpty = errs.Kind("empty")

	ErrURL     = errs.Entity("URL")
	ErrContent = errs.Entity("content")
)

var (
	ErrEmptyURL     = errs.New(errDomain, ErrEmpty, ErrURL)
	ErrEmptyContent = errs.New(errDomain, ErrEmpty, ErrContent)
)

const defaultTimeout = 15 * time.Second

type Config struct {
	id    string
	token string

	timeout time.Duration

	username  string
	avatarURL string
	tts       bool
}

func WithTimeout(dur time.Duration) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.timeout = dur

		return config
	})
}

func WithUsername(username string) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.username = username

		return config
	})
}

func WithAvatarURL(avatarURL string) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.avatarURL = avatarURL

		return config
	})
}

func WithTTS() cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.tts = true

		return config
	})
}
