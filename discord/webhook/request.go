package webhook

import (
	"github.com/zalgonoise/x/cfg"
)

type Request struct {
	username   string
	avatarURL  string
	tts        bool
	flags      uint64
	threadName string
}

func WithUsername(username string) cfg.Option[Request] {
	return cfg.Register(func(request Request) Request {
		request.username = username

		return request
	})
}

func WithAvatarURL(avatarURL string) cfg.Option[Request] {
	return cfg.Register(func(request Request) Request {
		request.avatarURL = avatarURL

		return request
	})
}

func WithTTS() cfg.Option[Request] {
	return cfg.Register(func(request Request) Request {
		request.tts = true

		return request
	})
}

func WithFlags(flags uint64) cfg.Option[Request] {
	return cfg.Register(func(request Request) Request {
		request.flags = flags

		return request
	})
}

func WithThreadName(threadName string) cfg.Option[Request] {
	return cfg.Register(func(request Request) Request {
		request.threadName = threadName

		return request
	})
}
