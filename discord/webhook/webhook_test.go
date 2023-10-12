package webhook_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zalgonoise/x/pluslog"
	"github.com/zalgonoise/x/pluslog/httplog"

	"github.com/zalgonoise/x/discord/log"
	"github.com/zalgonoise/x/discord/webhook"
)

func TestWebhook(t *testing.T) {
	logURL := os.Getenv("DISCORD_WEBHOOK_LOG_URL")
	postURL := os.Getenv("DISCORD_WEBHOOK_POST_URL")

	localLogger := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
	})

	remoteLogger := httplog.New(
		logURL, httplog.WithEncoder(log.JSON(true)),
	)

	logger := slog.New(pluslog.Multi(localLogger, remoteLogger))

	h, err := webhook.New(postURL, webhook.WithLogger(logger))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err = h.Execute(ctx, "beep boop this is a test message")
	require.NoError(t, err)
}
