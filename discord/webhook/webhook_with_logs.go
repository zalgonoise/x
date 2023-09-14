package webhook

import (
	"context"
	"github.com/switchupcb/dasgo/v10/dasgo"
	"io"
	"log/slog"
	"net/http"
)

type webhookWithLogs struct {
	w      Webhook
	logger *slog.Logger
}

func (w webhookWithLogs) Execute(ctx context.Context, content *dasgo.ExecuteWebhook) (res *http.Response, err error) {
	w.logger.InfoContext(ctx, "executing webhook", slog.Any("request", content))

	res, err = w.w.Execute(ctx, content)

	attrs := getResponseAttrs(res)

	if err != nil {
		w.logger.WarnContext(ctx, "webhook execution resulted in an error",
			slog.String("error", err.Error()),
			attrs,
		)

		return res, err
	}

	w.logger.InfoContext(ctx, "webhook execution succeeded", attrs)

	return res, nil
}

func getResponseAttrs(res *http.Response) slog.Attr {
	if res == nil {
		return slog.Attr{}
	}

	defer res.Body.Close()

	attrs := make([]any, 0, 2)
	attrs = append(attrs, slog.String("status", res.Status))

	body, err := io.ReadAll(res.Body)
	if err != nil {
		attrs = append(attrs, slog.Group("response", slog.String("body_read_error", err.Error())))

		return slog.Group("response", attrs...)
	}

	if len(body) > 0 {
		attrs = append(attrs, slog.String("body", string(body)))
	}

	return slog.Group("response", attrs...)
}

func withLogs(w Webhook, logger *slog.Logger) Webhook {
	if w == nil {
		return nil
	}

	if logger == nil {
		return w
	}

	return webhookWithLogs{
		w:      w,
		logger: logger,
	}
}
