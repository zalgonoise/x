package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/zalgonoise/x/cfg"
	"github.com/zalgonoise/x/errs"
	"net/http"
	"time"

	"github.com/switchupcb/dasgo/v10/dasgo"
)

const (
	baseURL                 = "https://discord.com/api"
	webhookExecuteURLFormat = "%s/webhooks/%s/%s" // /webhooks/{webhook.id}/{webhook.token}
	defaultTimeout          = 15 * time.Second

	errDomain = errs.Domain("x/discord/webhook")

	ErrEmpty = errs.Kind("empty")

	ErrID    = errs.Entity("ID")
	ErrToken = errs.Entity("token")
)

var (
	ErrEmptyID    = errs.New(errDomain, ErrEmpty, ErrID)
	ErrEmptyToken = errs.New(errDomain, ErrEmpty, ErrToken)
)

type Webhook interface {
	Execute(ctx context.Context, content *dasgo.ExecuteWebhook) (res *http.Response, err error)
}

type webhook struct {
	id    string
	token string

	timeout time.Duration
}

func (w webhook) Execute(ctx context.Context, content *dasgo.ExecuteWebhook) (res *http.Response, err error) {
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

func New(id, token string, options ...cfg.Option[Config]) (Webhook, error) {
	if id == "" {
		return noOpWebhook{}, ErrEmptyID
	}

	if token == "" {
		return noOpWebhook{}, ErrEmptyToken
	}

	config := cfg.New(options...)

	w := newWebhook(id, token, config)

	if config.logger != nil {
		w = withLogs(w, config.logger)
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

func (noOpWebhook) Execute(context.Context, *dasgo.ExecuteWebhook) (res *http.Response, err error) {
	return nil, nil
}
