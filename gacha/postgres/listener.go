package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zalgonoise/cfg"

	"github.com/zalgonoise/x/gacha/poller"
)

const (
	defaultReconnectBase  = 100 * time.Millisecond
	defaultReconnectLimit = 30 * time.Second
)

// ListenerConfig holds the configuration of a Listener.
type ListenerConfig struct {
	backoff func(retry int) time.Duration
	logger  *slog.Logger
}

func defaultListenerConfig() ListenerConfig {
	return ListenerConfig{
		backoff: poller.Exponential(defaultReconnectBase, defaultReconnectLimit),
		logger:  slog.New(slog.DiscardHandler),
	}
}

// WithReconnect sets the delay before the n-th consecutive reconnection attempt (n starts at 1) after the Listener
// loses its connection. It defaults to poller.Exponential(100*time.Millisecond, 30*time.Second).
func WithReconnect(backoff func(retry int) time.Duration) cfg.Option[ListenerConfig] {
	if backoff == nil {
		return cfg.NoOp[ListenerConfig]{}
	}

	return cfg.Register(func(c ListenerConfig) ListenerConfig {
		c.backoff = backoff

		return c
	})
}

// WithLogger sets the logger for connection failures. Nothing is logged by default.
func WithLogger(logger *slog.Logger) cfg.Option[ListenerConfig] {
	if logger == nil {
		return cfg.NoOp[ListenerConfig]{}
	}

	return cfg.Register(func(c ListenerConfig) ListenerConfig {
		c.logger = logger

		return c
	})
}

// Listener is a poller.Trigger fired by PostgreSQL notifications: it LISTENs on a channel, on a connection taken
// from a pool, and wakes the poller waiting on it whenever a notification arrives, such as those sent by a
// Repository created with WithNotify.
//
// Run must be running for the Listener to receive anything. Notifications are coalesced, like a poller.Notifier's.
// The Listener also fires once every time it (re)connects, since notifications sent while it was not listening are
// lost. Combine it with a poller.Ticker through poller.Any as a safety net for retries, whose backoff expires
// without any notification.
type Listener struct {
	pool     *pgxpool.Pool
	channel  string
	notifier *poller.Notifier
	config   ListenerConfig
}

// NewListener creates a Listener on channel, taking its connection from pool when it runs. It returns
// ErrInvalidChannel for an empty channel name.
func NewListener(pool *pgxpool.Pool, channel string, opts ...cfg.Option[ListenerConfig]) (*Listener, error) {
	if channel == "" {
		return nil, ErrInvalidChannel
	}

	return &Listener{
		pool:     pool,
		channel:  channel,
		notifier: poller.NewNotifier(),
		config:   cfg.Set(defaultListenerConfig(), opts...),
	}, nil
}

// Run listens for notifications until ctx is done, then returns nil. It takes one connection out of the pool for as
// long as it runs (closing it on return), so size the pool for it, and reconnects with backoff whenever the
// connection fails.
func (l *Listener) Run(ctx context.Context) error {
	for retry := 0; ; {
		connected, err := l.listen(ctx)

		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if connected {
			retry = 0
		}

		retry++

		l.config.logger.WarnContext(ctx, "gacha: listener disconnected",
			slog.String("channel", l.channel),
			slog.Int("retry", retry),
			slog.Any("error", err),
		)

		timer := time.NewTimer(l.config.backoff(retry))

		select {
		case <-ctx.Done():
			timer.Stop()

			return nil
		case <-timer.C:
		}
	}
}

// listen runs one connection's LISTEN loop. It reports whether it got to listen before failing.
func (l *Listener) listen(ctx context.Context) (bool, error) {
	pooled, err := l.pool.Acquire(ctx)
	if err != nil {
		return false, fmt.Errorf("%s: acquire listener connection: %w", domain, err)
	}

	// The connection leaves the pool: it would otherwise go back to it still listening.
	conn := pooled.Hijack()

	defer func() {
		closeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second)
		defer cancel()

		_ = conn.Close(closeCtx)
	}()

	if _, err = conn.Exec(ctx, "LISTEN "+pgx.Identifier{l.channel}.Sanitize()); err != nil {
		return false, fmt.Errorf("%s: listen: %w", domain, err)
	}

	// Catch up with whatever was written while not listening.
	l.notifier.Notify()

	for {
		if _, err = conn.WaitForNotification(ctx); err != nil {
			return true, fmt.Errorf("%s: wait for notification: %w", domain, err)
		}

		l.notifier.Notify()
	}
}

// Wait blocks until a notification arrives (or arrived since the last Wait), or ctx is done. It implements
// poller.Trigger.
func (l *Listener) Wait(ctx context.Context, last poller.Result) error {
	return l.notifier.Wait(ctx, last)
}

// Notify wakes the poller waiting on the Listener, as a notification would.
func (l *Listener) Notify() {
	l.notifier.Notify()
}
