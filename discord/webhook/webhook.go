package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/switchupcb/dasgo/v10/dasgo"
	"github.com/zalgonoise/x/cfg"
	"github.com/zalgonoise/x/errs"
)

const (
	baseURL                 = "https://discord.com/api/webhooks"
	webhookExecuteURLFormat = "%s/%s/%s" // /webhooks/{webhook.id}/{webhook.token}
	defaultTimeout          = 15 * time.Second

	errDomain = errs.Domain("x/discord/webhook")

	ErrEmpty = errs.Kind("empty")

	ErrURL = errs.Entity("URL")
)

var (
	ErrEmptyURL = errs.New(errDomain, ErrEmpty, ErrURL)
)

type Webhook interface {
	Execute(ctx context.Context, text string) (res *http.Response, err error)
	ExecuteContent(ctx context.Context, content *dasgo.ExecuteWebhook) (res *http.Response, err error)
}

type webhook struct {
	id    string
	token string

	timeout time.Duration
}

func (w webhook) Execute(ctx context.Context, text string) (res *http.Response, err error) {
	buf, err := json.Marshal(&dasgo.ExecuteWebhook{
		Content: &text,
	})
	if err != nil {
		return nil, err
	}

	client := http.Client{
		Transport: http.DefaultTransport,
		Timeout:   w.timeout,
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf(webhookExecuteURLFormat, baseURL, w.id, w.token),
		bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	res, err = client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode > 399 {
		return res, fmt.Errorf("HTTP request failed: status %s", res.Status)
	}

	return res, nil
}

func (w webhook) ExecuteContent(ctx context.Context, content *dasgo.ExecuteWebhook) (res *http.Response, err error) {
	buf, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}

	client := http.Client{
		Transport: http.DefaultTransport,
		Timeout:   w.timeout,
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf(webhookExecuteURLFormat, baseURL, w.id, w.token),
		bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	res, err = client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode > 399 {
		return res, fmt.Errorf("HTTP request failed: status %s", res.Status)
	}

	return res, nil
}

func Extract(url string) (id, token string, err error) {
	if url == "" {
		return "", "", ErrEmptyURL
	}

	split := strings.Split(strings.TrimPrefix(url, baseURL), "/")

	if len(split) == 3 && split[0] == "" &&
		split[1] != "" && split[2] != "" {
		return split[1], split[2], nil
	}

	return "", "", fmt.Errorf("invalid URL: %s", url)
}

func New(url string, options ...cfg.Option[Config]) (Webhook, error) {
	if url == "" {
		return noOpWebhook{}, ErrEmptyURL
	}

	id, token, err := Extract(url)
	if err != nil {
		return noOpWebhook{}, err
	}

	config := cfg.New(options...)

	w := newWebhook(id, token, config)

	if config.handler != nil {
		w = withLogs(w, config.handler)
	}

	return w, nil
}

func newWebhook(id, token string, config Config) Webhook {
	if config.timeout == 0 {
		config.timeout = defaultTimeout
	}

	return webhook{
		id:      id,
		token:   token,
		timeout: config.timeout,
	}
}

type noOpWebhook struct{}

func (noOpWebhook) Execute(context.Context, string) (res *http.Response, err error) {
	return nil, nil
}

func (noOpWebhook) ExecuteContent(context.Context, *dasgo.ExecuteWebhook) (res *http.Response, err error) {
	return nil, nil
}
