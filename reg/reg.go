package reg

import (
	"context"
	"log/slog"
	"os"

	"github.com/zalgonoise/cfg"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Registrar struct {
	logger     *slog.Logger
	traceAttrs []attribute.KeyValue
	logAttrs   []any
}

func New(logger *slog.Logger, traceAttrs []attribute.KeyValue, logAttrs []any) *Registrar {
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))
	}

	return &Registrar{
		logger:     logger,
		traceAttrs: traceAttrs,
		logAttrs:   logAttrs,
	}
}

type Event struct {
	ctx        context.Context
	err        error
	message    string
	span       trace.Span
	traceAttrs []attribute.KeyValue
	logger     *slog.Logger
	logAttrs   []any
	logLevel   slog.Level
	metric     func()
}

func (r *Registrar) Event(ctx context.Context, message string, opts ...cfg.Option[*Event]) {
	e := cfg.Set(&Event{ctx: ctx, message: message, logger: r.logger, logLevel: slog.LevelError}, opts...)

	e.logAttrs = append(e.logAttrs, r.logAttrs...)
	e.traceAttrs = append(e.traceAttrs, r.traceAttrs...)

	if e.err == nil {
		e.logger.Log(ctx, e.logLevel, e.message, e.logAttrs...)

		if e.span != nil {
			e.span.AddEvent(e.message, trace.WithAttributes(e.traceAttrs...))
		}

		if e.metric != nil {
			e.metric()
		}

		return
	}

	// log
	e.logger.Log(ctx, e.logLevel, e.message, append(e.logAttrs, slog.String("error", e.err.Error()))...)

	// trace
	if e.span != nil {
		e.span.SetStatus(codes.Error, e.err.Error())
		e.span.RecordError(e.err)
		e.span.AddEvent(e.message, trace.WithAttributes(append(e.traceAttrs, attribute.String("error", e.err.Error()))...))
	}

	//metric
	if e.metric != nil {
		e.metric()
	}
}

func WithError(err error) cfg.Option[*Event] {
	if err == nil {
		return cfg.Register(func(e *Event) *Event {
			e.logLevel = slog.LevelInfo

			return e
		})
	}

	return cfg.Register(func(e *Event) *Event {
		e.logLevel = slog.LevelError
		e.err = err

		return e
	})
}

func WithSpan(span trace.Span, attrs ...attribute.KeyValue) cfg.Option[*Event] {
	if span == nil {
		return cfg.NoOp[*Event]{}
	}

	return cfg.Register(func(e *Event) *Event {
		e.span = span

		if len(attrs) > 0 {
			e.traceAttrs = append(e.traceAttrs, attrs...)
		}

		return e
	})
}

func WithLogAttributes(attrs ...slog.Attr) cfg.Option[*Event] {
	if len(attrs) == 0 {
		return cfg.NoOp[*Event]{}
	}

	logAttrs := make([]any, len(attrs))
	for i, attr := range attrs {
		logAttrs[i] = attr
	}

	return cfg.Register(func(e *Event) *Event {
		e.logAttrs = append(e.logAttrs, logAttrs...)

		return e
	})
}

func WithLogLevel(level slog.Level) cfg.Option[*Event] {
	return cfg.Register(func(e *Event) *Event {
		e.logLevel = level

		return e
	})
}

func WithMetric(metric func()) cfg.Option[*Event] {
	if metric == nil {
		return cfg.NoOp[*Event]{}
	}

	return cfg.Register(func(e *Event) *Event {
		e.metric = metric

		return e
	})
}
