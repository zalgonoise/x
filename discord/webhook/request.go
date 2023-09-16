package webhook

import (
	"github.com/switchupcb/dasgo/v10/dasgo"
	"github.com/zalgonoise/x/cfg"
)

func Content(content string, options ...cfg.Option[Request]) *dasgo.ExecuteWebhook {
	if content == "" {
		return nil
	}

	req := cfg.New(options...)
	h := newExecuteWebhook(req)

	h.Content = &content

	return h
}

func newExecuteWebhook(req Request) *dasgo.ExecuteWebhook {
	var (
		username   *string
		avatarURL  *string
		tts        *bool
		flags      *dasgo.BitFlag
		threadName *string
	)

	if req.tts {
		tts = &req.tts
	}

	if req.username != "" {
		username = &req.username
	}

	if req.avatarURL != "" {
		avatarURL = &req.avatarURL
	}

	if req.flags != 0 {
		flags = (*dasgo.BitFlag)(&req.flags)
	}

	if req.threadName != "" {
		threadName = &req.threadName
	}

	return &dasgo.ExecuteWebhook{
		Username:   username,
		AvatarURL:  avatarURL,
		TTS:        tts,
		Flags:      flags,
		ThreadName: threadName,
	}
}

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
