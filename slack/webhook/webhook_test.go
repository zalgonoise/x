package webhook_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zalgonoise/x/slack/webhook"
)

func TestWebhook(t *testing.T) {
	url := os.Getenv("SLACK_WEBHOOK_POST_URL")
	_, _, _, err := webhook.Extract(url)
	require.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
	}))

	h, err := webhook.New(url, webhook.WithLogger(logger))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err = h.Execute(ctx, "beep boop~ some text")
	require.NoError(t, err)
}
