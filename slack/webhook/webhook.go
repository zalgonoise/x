package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/zalgonoise/x/cfg"
	"github.com/zalgonoise/x/errs"
)

const (
	baseURL                 = "https://hooks.slack.com/services"
	webhookExecuteURLFormat = "%s/%s/%s/%s" // /services/{webhook.team}/{webhook.channel}/{webhook.token}
	defaultTimeout          = 15 * time.Second

	errDomain = errs.Domain("x/discord/webhook")

	ErrEmpty = errs.Kind("empty")

	ErrURL = errs.Entity("URL")
)

var (
	ErrEmptyURL = errs.WithDomain(errDomain, ErrEmpty, ErrURL)
)

type Webhook interface {
	Execute(ctx context.Context, text string) (res *http.Response, err error)
}

type webhook struct {
	team    string
	channel string
	token   string

	timeout time.Duration
}

type SlackWebhook struct {
	Channel *string          `json:"channel,omitempty"`
	Text    *string          `json:"text,omitempty"`
	Blocks  []map[string]any `json:"blocks,omitempty"`
}

func (w webhook) Execute(ctx context.Context, text string) (res *http.Response, err error) {
	buf, err := json.Marshal(SlackWebhook{
		Text: &text,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf(webhookExecuteURLFormat, baseURL, w.team, w.channel, w.token),
		bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	res, err = (&http.Client{
		Transport: http.DefaultTransport,
		Timeout:   w.timeout,
	}).Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode > 399 {
		return res, fmt.Errorf("HTTP request failed: status %s", res.Status)
	}

	return res, nil
}

func Extract(url string) (team, channel, token string, err error) {
	if url == "" {
		return "", "", "", ErrEmptyURL
	}

	split := strings.Split(strings.TrimPrefix(url, baseURL), "/")

	if len(split) == 4 && split[0] == "" &&
		split[1] != "" && split[2] != "" && split[3] != "" {
		return split[1], split[2], split[3], nil
	}

	return "", "", "", fmt.Errorf("invalid URL: %s", url)
}

func New(url string, options ...cfg.Option[Config]) (Webhook, error) {
	if url == "" {
		return noOpWebhook{}, ErrEmptyURL
	}

	config := cfg.New(options...)

	w, err := newWebhook(url, config)
	if err != nil {
		return noOpWebhook{}, err
	}

	if config.handler != nil {
		w = withLogs(w, config.handler)
	}

	return w, nil
}

func newWebhook(url string, config Config) (Webhook, error) {
	if config.timeout == 0 {
		config.timeout = defaultTimeout
	}

	team, channel, token, err := Extract(url)
	if err != nil {
		return noOpWebhook{}, err
	}

	return webhook{
		team:    team,
		channel: channel,
		token:   token,
		timeout: config.timeout,
	}, nil
}

type noOpWebhook struct{}

func (noOpWebhook) Execute(context.Context, string) (res *http.Response, err error) {
	return nil, nil
}
