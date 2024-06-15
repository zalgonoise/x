package events

import (
	"context"
	"log/slog"
	"strings"

	"github.com/zalgonoise/x/discord/webhook"
)

type Event struct {
	Action string
	URI    string
	Module string
	Branch string
	Output []string
}

type Reporter struct {
	logger  *slog.Logger
	webhook webhook.Webhook
}

func (r *Reporter) ReportEvent(ctx context.Context, event Event) {
	go func() {
		sb := &strings.Builder{}

		sb.WriteString("Action: ")
		sb.WriteString(event.Action)
		sb.WriteString("\nRepository: https://")
		sb.WriteString(event.URI)
		if event.Branch != "" {
			sb.WriteString("/tree/")
			sb.WriteString(event.Branch)
		}

		if event.Module != "" {
			sb.WriteByte('/')
			sb.WriteString(event.Module)
		}

		if len(event.Output) > 0 {
			sb.WriteString("\n\n")
		}

		for i := range event.Output {
			if i > 0 {
				sb.WriteByte('\n')
			}

			sb.WriteString("- ")
			sb.WriteString(event.Output[i])
		}

		res, err := r.webhook.Execute(ctx, sb.String())
		if err != nil {
			r.logger.WarnContext(ctx, "failed to issue event via webhook",
				slog.String("error", err.Error()))
		}

		if res != nil && res.Body != nil {
			_ = res.Body.Close()
		}
	}()
}

func NewReporter(uri string, logger *slog.Logger) (*Reporter, error) {
	w, err := webhook.New(uri)
	if err != nil {
		return nil, err
	}

	return &Reporter{
		logger:  logger,
		webhook: w,
	}, nil
}
