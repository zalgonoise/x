package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/switchupcb/dasgo/v10/dasgo"
	"github.com/zalgonoise/x/cfg"
)

const (
	baseURL                 = "https://discord.com/api"
	webhookExecuteURLFormat = "%s/webhooks/%s/%s" // /webhooks/{webhook.id}/{webhook.token}
)

type Webhook struct {
	id    string
	token string

	config Config
}

func New(id, token string, options ...cfg.Option[Config]) (Webhook, error) {
	config := cfg.New(options...)
	if config.timeout == 0 {
		config.timeout = defaultTimeout
	}

	return Webhook{
		id:     id,
		token:  token,
		config: config,
	}, nil
}

func (w Webhook) Execute(ctx context.Context, content string) error {
	if content == "" {
		return ErrEmptyContent
	}

	var (
		username  *string
		avatarURL *string
		tts       *bool
	)

	if w.config.username != "" {
		username = &w.config.username
	}

	if w.config.avatarURL != "" {
		avatarURL = &w.config.avatarURL
	}

	if w.config.tts {
		tts = &w.config.tts
	}

	h := dasgo.ExecuteWebhook{
		Content:   &content,
		Username:  username,
		AvatarURL: avatarURL,
		TTS:       tts,
	}

	buf, err := json.Marshal(h)
	if err != nil {
		return err
	}

	client := http.Client{
		Transport: http.DefaultTransport,
		Timeout:   w.config.timeout,
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf(webhookExecuteURLFormat, baseURL, w.id, w.token),
		bytes.NewReader(buf))
	if err != nil {
		return err
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode > 399 {
		body, readErr := io.ReadAll(res.Body)
		if readErr != nil {
			return fmt.Errorf("HTTP request failed: status %s; body read error: %w", res.Status, err)
		}

		return fmt.Errorf("HTTP request failed: status %s; body: %s", res.Status, string(body))
	}

	return nil
}
