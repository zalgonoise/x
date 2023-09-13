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

func Embed(embeds []*dasgo.Embed, options ...cfg.Option[Request]) *dasgo.ExecuteWebhook {
	if len(embeds) == 0 {
		return nil
	}

	req := cfg.New(options...)
	h := newExecuteWebhook(req)

	h.Embeds = embeds

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
