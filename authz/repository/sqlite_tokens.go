package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/micron"
	"github.com/zalgonoise/micron/executor"
	"github.com/zalgonoise/micron/schedule"
	"github.com/zalgonoise/micron/selector"
	"go.opentelemetry.io/otel/trace"
)

const (
	queryChallengesCreate = `
INSERT INTO challenges (service_id, challenge, expiry)
VALUES ((SELECT id FROM services WHERE name = ?), ?, ?)
`

	queryChallengesGet = `
SELECT challenge, expiry FROM challenges
WHERE service_id = (
	SELECT id FROM services WHERE name = ?
)
`

	queryChallengesDelete = `
DELETE FROM challenges WHERE service_id = (
	SELECT id FROM services WHERE name = ?
)
`

	queryChallengesCleanup = `
DELETE FROM challenges WHERE expiry < ?
`

	queryTokensCreate = `
INSERT INTO tokens (service_id, token, expiry)
VALUES ((SELECT id FROM services WHERE name = ?), ?, ?)
`

	queryTokensGet = `
SELECT token, expiry FROM tokens
WHERE service_id = (
	SELECT id FROM services WHERE name = ?
)
`

	queryTokensDelete = `
DELETE FROM tokens WHERE service_id = (
	SELECT id FROM services WHERE name = ?
)
`

	queryTokensCleanup = `
DELETE FROM tokens WHERE WHERE expiry < ?
`
)

type Tokens struct {
	cleanupTimeout  time.Duration
	cleanupSchedule string

	db *sql.DB

	done context.CancelFunc

	logger *slog.Logger
	m      Metrics
	tracer trace.Tracer
}

func NewTokens(db *sql.DB, opts ...cfg.Option[Config]) (*Tokens, error) {
	config := cfg.Set(defaultConfig(), opts...)

	if config.cleanupTimeout <= 0 {
		config.cleanupTimeout = defaultCleanupTimeout
	}

	if config.cleanupSchedule == "" {
		config.cleanupSchedule = defaultCleanupSchedule
	}

	ctx, done := context.WithCancel(context.Background())

	ca := &Tokens{
		cleanupTimeout:  config.cleanupTimeout,
		cleanupSchedule: config.cleanupSchedule,
		db:              db,
		done:            done,
		logger:          config.logger,
		m:               config.m,
		tracer:          config.tracer,
	}

	if err := ca.runCron(ctx); err != nil {
		return nil, err
	}

	return ca, nil
}

func (r *Tokens) GetChallenge(ctx context.Context, service string) (challenge []byte, expiry time.Time, err error) {
	var expiryMillis int64

	if err = r.db.QueryRowContext(ctx, queryChallengesGet, service).Scan(&challenge, &expiryMillis); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, time.Time{}, ErrNotFound
		}

		return nil, time.Time{}, err
	}

	return challenge, time.UnixMilli(expiryMillis), nil
}

func (r *Tokens) CreateChallenge(
	ctx context.Context, service string, challenge []byte, expiry time.Time) error {
	res, err := r.db.ExecContext(ctx, queryChallengesCreate, service, challenge, expiry.UnixMilli())
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows != 1 {
		return fmt.Errorf("%w: %s", ErrFailedDBWrite, service)
	}

	return nil
}

func (r *Tokens) DeleteChallenge(ctx context.Context, service string) error {
	var (
		pubKey       []byte
		expiryMillis int64
	)

	if err := r.db.QueryRowContext(ctx, queryChallengesGet, service).Scan(&pubKey, &expiryMillis); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		return err
	}

	res, err := r.db.ExecContext(ctx, queryChallengesDelete, service)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows != 1 {
		return fmt.Errorf("%w: %s", ErrFailedDBDelete, service)
	}

	return nil
}

func (r *Tokens) GetToken(ctx context.Context, service string) (token []byte, expiry time.Time, err error) {
	var expiryMillis int64

	if err = r.db.QueryRowContext(ctx, queryTokensGet, service).Scan(&token, &expiryMillis); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, time.Time{}, ErrNotFound
		}

		return nil, time.Time{}, err
	}

	return token, time.UnixMilli(expiryMillis), nil
}

func (r *Tokens) CreateToken(ctx context.Context, service string, token []byte, expiry time.Time) error {
	res, err := r.db.ExecContext(ctx, queryTokensCreate, service, token, expiry.UnixMilli())
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows != 1 {
		return fmt.Errorf("%w: %s", ErrFailedDBWrite, service)
	}

	return nil
}

func (r *Tokens) DeleteToken(ctx context.Context, service string) error {
	var (
		pubKey       []byte
		expiryMillis int64
	)

	if err := r.db.QueryRowContext(ctx, queryTokensGet, service).Scan(&pubKey, &expiryMillis); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		return err
	}

	res, err := r.db.ExecContext(ctx, queryTokensDelete, service)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows != 1 {
		return fmt.Errorf("%w: %s", ErrFailedDBDelete, service)
	}

	return nil
}

func (r *Tokens) Close() error {
	if r.done != nil {
		r.done()
	}

	return r.db.Close()
}

func (r *Tokens) cleanup(ctx context.Context) error {
	ctx, done := context.WithTimeout(context.Background(), r.cleanupTimeout)
	defer done()

	if _, err := r.db.ExecContext(ctx, queryTokensCleanup, time.Now()); err != nil {
		return err
	}

	_, err := r.db.ExecContext(ctx, queryChallengesCleanup, time.Now())

	return err
}

func (r *Tokens) runCron(ctx context.Context) error {
	s, err := schedule.New(
		schedule.WithSchedule(r.cleanupSchedule),
		schedule.WithLogger(r.logger),
		schedule.WithTrace(r.tracer),
		schedule.WithMetrics(r.m),
	)
	if err != nil {
		return err
	}

	exec, err := executor.New("tokens_cleanup",
		executor.WithScheduler(s),
		executor.WithRunners(executor.Runnable(r.cleanup)),
		executor.WithLocation(time.UTC),
		executor.WithLogger(r.logger),
		executor.WithTrace(r.tracer),
		executor.WithMetrics(r.m),
	)
	if err != nil {
		return err
	}

	sel, err := selector.New(
		selector.WithExecutors(exec),
		selector.WithTimeout(time.Minute),
		selector.WithLogger(r.logger),
		selector.WithTrace(r.tracer),
		selector.WithMetrics(r.m),
	)
	if err != nil {
		return err
	}

	cron, err := micron.New(
		micron.WithSelector(sel),
		micron.WithErrorBufferSize(16),
		micron.WithLogger(r.logger),
		micron.WithTrace(r.tracer),
	)
	if err != nil {
		return err
	}

	go cron.Run(ctx)

	return nil
}