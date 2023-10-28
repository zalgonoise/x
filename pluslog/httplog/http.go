package httplog

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"time"

	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/x/errs"
)

const (
	defaultTimeout = 15 * time.Second

	errDomain = errs.Domain("x/pluslog/httplog")

	ErrFailed = errs.Kind("failed")

	ErrHTTPRequest = errs.Entity("HTTP request")
)

var (
	ErrFailedHTTPRequest = errs.New(errDomain, ErrFailed, ErrHTTPRequest)
)

type Encoder interface {
	Encode(HTTPRecord) ([]byte, error)
}

type HTTPLogger struct {
	config Config

	url   string
	attrs []slog.Attr
	group string
}

func New(url string, options ...cfg.Option[Config]) slog.Handler {
	if url == "" {
		return nil
	}

	config := cfg.New(options...)

	if config.encoder == nil {
		config.encoder = MarshalJSON{}
	}

	if config.level == nil {
		config.level = slog.LevelDebug
	}

	if config.timeout == 0 {
		config.timeout = defaultTimeout
	}

	return HTTPLogger{
		config: config,
		url:    url,
	}
}

// Enabled reports whether the handler handles records at the given level.
// The handler ignores records whose level is lower.
// It is called early, before any arguments are processed,
// to save effort if the log event should be discarded.
// If called from a Logger method, the first argument is the context
// passed to that method, or context.Background() if nil was passed
// or the method does not take a context.
// The context is passed so Enabled can use its values
// to make a decision.
func (h HTTPLogger) Enabled(_ context.Context, level slog.Level) bool {
	return level > h.config.level.Level()
}

// Handle handles the Record.
// It will only be called when Enabled returns true.
// The Context argument is as for Enabled.
// It is present solely to provide Handlers access to the context's values.
// Canceling the context should not affect record processing.
// (Among other things, log messages may be necessary to debug a
// cancellation-related problem.)
//
// Handle methods that produce output should observe the following rules:
//   - If r.Time is the zero time, ignore the time.
//   - If r.PC is zero, ignore it.
//   - Attr's values should be resolved.
//   - If an Attr's key and value are both the zero value, ignore the Attr.
//     This can be tested with attr.Equal(Attr{}).
//   - If a group's key is empty, inline the group's Attrs.
//   - If a group has no Attrs (even if it has a non-empty key),
//     ignore it.
//
//nolint:gocritic // this method implements the slog.Handler interface
func (h HTTPLogger) Handle(ctx context.Context, record slog.Record) error {
	handlerAttrs := mapAttrs(h.attrs)
	attrs := extractAttrs(handlerAttrs, record)

	if h.group != "" {
		attrs = map[string]any{
			h.group: attrs,
		}
	}

	var src *slog.Source
	if h.config.source {
		src = source(record.PC)
	}

	r := HTTPRecord{
		Time:    record.Time,
		Message: record.Message,
		Level:   record.Level.String(),
		Source:  src,
		Attrs:   attrs,
	}

	buf, err := h.config.encoder.Encode(r)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		h.url,
		bytes.NewReader(buf),
	)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{
		Timeout: h.config.timeout,
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode > 399 {
		return fmt.Errorf("%w: status: %s", ErrFailedHTTPRequest, res.Status)
	}

	return nil
}

// WithAttrs returns a new Handler whose attributes consist of
// both the receiver's attributes and the arguments.
// The Handler owns the slice: it may retain, modify or discard it.
func (h HTTPLogger) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	if len(h.attrs) == 0 {
		return &HTTPLogger{
			config: h.config,
			url:    h.url,
			attrs:  attrs,
			group:  h.group,
		}
	}

	oldAttrs := make([]slog.Attr, 0, len(h.attrs))

oldAttrLoop:
	for i := range h.attrs {
		for idx := range attrs {
			if attrs[idx].Key == h.attrs[i].Key {
				continue oldAttrLoop
			}
		}

		oldAttrs = append(oldAttrs, h.attrs[i])
	}

	return &HTTPLogger{
		config: h.config,
		url:    h.url,
		attrs:  append(oldAttrs, attrs...),
		group:  h.group,
	}
}

// WithGroup returns a new Handler with the given group appended to
// the receiver's existing groups.
// The keys of all subsequent attributes, whether added by With or in a
// Record, should be qualified by the sequence of group names.
//
// How this qualification happens is up to the Handler, so long as
// this Handler's attribute keys differ from those of another Handler
// with a different sequence of group names.
//
// A Handler should treat WithGroup as starting a Group of Attrs that ends
// at the end of the log event. That is,
//
//	logger.WithGroup("s").LogAttrs(level, msg, slog.Int("a", 1), slog.Int("b", 2))
//
// should behave like
//
//	logger.LogAttrs(level, msg, slog.Group("s", slog.Int("a", 1), slog.Int("b", 2)))
//
// If the name is empty, WithGroup returns the receiver.
func (h HTTPLogger) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	return &HTTPLogger{
		config: h.config,
		url:    h.url,
		attrs:  h.attrs,
		group:  name,
	}
}

func mapAttrs(attrs []slog.Attr) map[string]any {
	m := make(map[string]any, len(attrs))

	for i := range attrs {
		_ = mapValue(m, attrs[i].Key, attrs[i].Value)
	}

	if len(m) == 0 {
		return nil
	}

	return m
}

func extractAttrs(attrs map[string]any, record slog.Record) map[string]any {
	if attrs == nil {
		attrs = make(map[string]any, record.NumAttrs())
	}

	record.Attrs(func(attr slog.Attr) bool {
		return mapValue(attrs, attr.Key, attr.Value)
	})

	if len(attrs) == 0 {
		return nil
	}

	return attrs
}

func mapValue(attrs map[string]any, key string, value slog.Value) bool {
	switch value.Kind() {
	case slog.KindAny:
		attrs[key] = value.Any()
	case slog.KindBool:
		attrs[key] = value.Bool()
	case slog.KindDuration:
		attrs[key] = value.Duration().String()
	case slog.KindFloat64:
		attrs[key] = value.Float64()
	case slog.KindInt64:
		attrs[key] = value.Int64()
	case slog.KindString:
		attrs[key] = value.String()
	case slog.KindTime:
		attrs[key] = value.Time().Format(time.RFC3339)
	case slog.KindUint64:
		attrs[key] = value.Uint64()
	case slog.KindGroup:
		groupAttrs := value.Group()
		inner := make(map[string]any, len(groupAttrs))

		for i := range groupAttrs {
			mapValue(inner, groupAttrs[i].Key, groupAttrs[i].Value)
		}

		attrs[key] = inner
	case slog.KindLogValuer:
		mapValue(attrs, key, value.LogValuer().LogValue())
	default:
		return false
	}

	return true
}

// source returns a Source for the log event.
// If the Record was created without the necessary information,
// or if the location is unavailable, it returns a non-nil *Source
// with zero fields.
func source(pc uintptr) *slog.Source {
	fs := runtime.CallersFrames([]uintptr{pc})
	f, _ := fs.Next()
	return &slog.Source{
		Function: f.Function,
		File:     f.File,
		Line:     f.Line,
	}
}
