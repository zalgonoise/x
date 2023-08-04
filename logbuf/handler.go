package logbuf

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/otel/trace"
	"golang.org/x/exp/slog"
)

const minAlloc = 64

// Repository describes the actions that the trace ID store should contain
type Repository interface {
	// Shutdown gracefully stops the Repository's underlying sql.DB
	Shutdown(ctx context.Context) error
	// InsertTrace adds the input trace.TraceID to the database if it does not yet exist, alongside with the current
	// timestamp (of when it is registered). Returns an error if raised.
	InsertTrace(ctx context.Context, traceID trace.TraceID) (err error)
	// DeleteTraces removes all trace.TraceID from the database that are older than the threshold time.Duration
	// (which is calculated from the current time minus this value). It returns a slice of trace.TraceID with all
	// values that are were removed and an error if raised.
	DeleteTraces(ctx context.Context, threshold time.Duration) (pruned []trace.TraceID, err error)
}

// BufferedHandler will store the incoming slog.Records in-memory in a map, identified by their trace ID.
//
// This data type is connected to a SQLite instance (in-memory or otherwise) that will store the trace ID of a
// slog.Record through the provided context (if does not exist), alongside its timestamp. All subsequent slog.Record
// are appended to the slice of slog.Record under its trace ID.
//
// In the background, the BufferedHandler will periodically scan the database for trace IDs that have been stored for
// too long, and prune those slog.Record from both the map and its trace ID from the database. This is done by scanning
// the database for trace IDs older than the configured duration and using that trace ID as reference to remove the
// slog.Record data from the map.
//
// If an incoming slog.Record contains a level greater or equal to the configured threshold slog.Level, then all the
// slog.Record entries in the map corresponding to its trace ID are flushed to the slog.Handler that BufferedHandler is
// wrapping.
//
// Besides the slog.Handler implementation, BufferedHandler also exposes a Shutdown method to gracefully stop the
// handler.
//
// NOTE: working with a BufferedHandler does not guarantee slog.Record persistence as it is based on an ephemeral,
// in-memory model. This means that all the information stored in this data structure is lost if, for example, the
// application crashes with a panic.
type BufferedHandler struct {
	cfg *BufferedHandlerConfig

	h slog.Handler

	done       func()
	repository Repository
	cache      map[trace.TraceID][]slog.Record
}

// NewBufferedHandler creates a BufferedHandler with the input BufferedHandlerConfig, slog.Handler and Repository.
func NewBufferedHandler(
	ctx context.Context, cfg *BufferedHandlerConfig, handler slog.Handler, repo Repository,
) (*BufferedHandler, error) {
	h := &BufferedHandler{
		cfg:        cfg,
		h:          handler,
		repository: repo,
		cache:      make(map[trace.TraceID][]slog.Record),
	}

	ctx, h.done = context.WithCancel(ctx)

	go h.run(ctx)

	return h, nil
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
func (h *BufferedHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.h.Enabled(ctx, level)
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
func (h *BufferedHandler) Handle(ctx context.Context, r slog.Record) error {
	traceID, err := GetTraceID(ctx)
	if err != nil {
		return err
	}

	if r.Level >= slog.Level(h.cfg.FlushLevel) {
		if logs, ok := h.cache[traceID]; ok {
			errs := make([]error, 0, len(logs))

			for i := range logs {
				if err = h.h.Handle(ctx, logs[i]); err != nil {
					errs = append(errs, err)
				}
			}

			if err = h.h.Handle(ctx, r); err != nil {
				errs = append(errs, err)
			}

			switch len(errs) {
			case 0:
				delete(h.cache, traceID)

				return nil
			case 1:
				return errs[0]
			default:
				return errors.Join(errs...)
			}
		}

	}

	if _, ok := h.cache[traceID]; !ok {
		h.cache[traceID] = make([]slog.Record, 0, minAlloc)
		h.cache[traceID] = append(h.cache[traceID], r)

		if err = h.repository.InsertTrace(ctx, traceID); err != nil {
			return err
		}

		return nil
	}

	h.cache[traceID] = append(h.cache[traceID], r)

	return nil
}

// WithAttrs returns a new Handler whose attributes consist of
// both the receiver's attributes and the arguments.
// The Handler owns the slice: it may retain, modify or discard it.
func (h *BufferedHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.h.WithAttrs(attrs)

	return h
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
func (h *BufferedHandler) WithGroup(name string) slog.Handler {
	h.h.WithGroup(name)

	return h
}

// Shutdown gracefully stops the BufferedHandler, returning an error if raised
func (h *BufferedHandler) Shutdown(ctx context.Context) error {
	h.done()

	return h.repository.Shutdown(ctx)
}

// run periodically scans the database for expired trace IDs, pruning them from both the database and the slog.Record
// cache. This function should be non-blocking, to be executed as a goroutine.
func (h *BufferedHandler) run(ctx context.Context) {
	var (
		ticker = time.NewTicker(h.cfg.FlushFrequency)
		sigCh  = make(chan os.Signal, 1)
	)
	defer ticker.Stop()

	signal.Notify(sigCh, os.Interrupt, os.Kill, syscall.SIGTERM)

	for {
		select {
		case <-ctx.Done():
			return
		case <-sigCh:
			return
		case _ = <-ticker.C:
			// don't remove logs from cache on DB error
			// needs a different approach
			if traceIDs, err := h.repository.DeleteTraces(ctx, h.cfg.DeletionThresh); err == nil {
				for i := range traceIDs {
					delete(h.cache, traceIDs[i])
				}
			}
		}
	}
}
