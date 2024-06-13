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

		sb.WriteString(`New event for https://`)
		sb.WriteString(event.URI)
		sb.WriteString("/tree/")
		sb.WriteString(event.Branch)
		sb.WriteByte('/')
		sb.WriteString(event.Module)
		sb.WriteString("!")

		if len(event.Output) > 0 {
			sb.WriteString("\n\n")
		}

		for i := range event.Output {
			if i > 0 {
				sb.WriteByte('\n')
			}

			sb.WriteString("\t- ")
			sb.WriteString(event.Output[i])
		}

		res, err := r.webhook.Execute(ctx, sb.String())
		if err != nil {
			r.logger.WarnContext(ctx, "failed to issue event via webhook",
				slog.String("error", err.Error()))
		}

		_ = res.Body.Close()
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
