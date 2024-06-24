package events

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/zalgonoise/x/discord/webhook"
)

const minBufferSize = 8

type Event struct {
	Action string
	URI    string
	Module string
	Branch string
	Output []string
}

type Reporter struct {
	logger  *slog.Logger
	buf     *buffer
	webhook webhook.Webhook
}

func (r *Reporter) ReportEvent(event Event) {
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

	r.buf.WriteString(sb.String())
}

func (r *Reporter) Flush() {
	r.buf.flush(r.buf.String())
}

func (r *Reporter) flush(s string) {
	if s == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	res, err := r.webhook.Execute(ctx, s)
	if err != nil {
		r.logger.WarnContext(ctx, "failed to issue event via webhook",
			slog.String("error", err.Error()))
	}

	if res != nil && res.Body != nil {
		_ = res.Body.Close()
	}
}

type buffer struct {
	flush func(s string)

	i      int
	values []string
	mu     *sync.RWMutex
}

func (b *buffer) WriteString(s string) {
	b.mu.Lock()
	b.values[b.i] = s
	b.i++
	b.mu.Unlock()

	if b.i == cap(b.values) {
		b.flush(b.String())
	}
}

func (b *buffer) String() string {
	v := make([]string, 0, b.i)

	b.mu.RLock()
	copy(v, b.values[:b.i])
	b.i = 0
	b.mu.RUnlock()

	return strings.Join(v, "\n")
}

func NewReporter(uri string, bufferSize int, logger *slog.Logger) (*Reporter, error) {
	if bufferSize <= 0 {
		bufferSize = minBufferSize
	}

	w, err := webhook.New(uri)
	if err != nil {
		return nil, err
	}

	r := &Reporter{
		logger:  logger,
		webhook: w,
	}

	buf := &buffer{
		flush:  r.flush,
		values: make([]string, bufferSize),
		mu:     &sync.RWMutex{},
	}

	r.buf = buf

	return r, nil
}
