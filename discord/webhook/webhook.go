package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/switchupcb/dasgo/v10/dasgo"
)

const (
	baseURL                 = "https://discord.com/api"
	webhookExecuteURLFormat = "%s/webhooks/%s/%s" // /webhooks/{webhook.id}/{webhook.token}
	defaultTimeout          = 15 * time.Second
)

type Webhook struct {
	id    string
	token string

	timeout time.Duration
}

func New(id, token string, timeout time.Duration) (Webhook, error) {
	if timeout == 0 {
		timeout = defaultTimeout
	}

	return Webhook{
		id:      id,
		token:   token,
		timeout: timeout,
	}, nil
}

func (w Webhook) Execute(ctx context.Context, content *dasgo.ExecuteWebhook) (data []byte, err error) {
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

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		return nil, fmt.Errorf("HTTP request failed: status %s; body read error: %w", res.Status, err)
	}

	if res.StatusCode > 399 {
		return body, fmt.Errorf("HTTP request failed: status %s; body: %s", res.Status, string(body))
	}

	return body, nil
}
