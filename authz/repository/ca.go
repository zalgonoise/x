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
	queryGet = `
SELECT pub_key, cert FROM services
WHERE name = ?
`

	queryCreate = `
INSERT INTO services (name, pub_key, cert, expiry)
VALUES (?, ?, ?)
`

	queryDelete = `
DELETE FROM services WHERE name = ?
`

	queryCleanup = `
DELETE FROM services WHERE name IN (
    SELECT name FROM services 
    WHERE expiry < ?
)
`
)

var (
	ErrNotFound       = errors.New("not found")
	ErrFailedDBWrite  = errors.New("failed to write entry")
	ErrFailedDBDelete = errors.New("failed to remove entry")
)

type Metrics interface {
	IncSchedulerNextCalls()
	IncExecutorExecCalls(id string)
	IncExecutorExecErrors(id string)
	ObserveExecLatency(ctx context.Context, id string, dur time.Duration)
	IncExecutorNextCalls(id string)
	IncSelectorSelectCalls()
	IncSelectorSelectErrors()
}

type CertificateAuthority struct {
	cleanupTimeout  time.Duration
	cleanupSchedule string

	db *sql.DB

	done context.CancelFunc

	logger *slog.Logger
	m      Metrics
	tracer trace.Tracer
}

func NewCertificateAuthority(db *sql.DB, opts ...cfg.Option[Config]) (*CertificateAuthority, error) {
	config := cfg.Set(defaultConfig(), opts...)

	if config.cleanupTimeout <= 0 {
		config.cleanupTimeout = defaultCleanupTimeout
	}

	if config.cleanupSchedule == "" {
		config.cleanupSchedule = defaultCleanupSchedule
	}

	ctx, done := context.WithCancel(context.Background())

	ca := &CertificateAuthority{
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

func (r *CertificateAuthority) Get(ctx context.Context, service string) (pubKey []byte, cert []byte, err error) {
	if err = r.db.QueryRowContext(ctx, queryGet, service).Scan(&pubKey, &cert); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, ErrNotFound
		}

		return nil, nil, err
	}

	return pubKey, cert, nil
}

func (r *CertificateAuthority) Create(
	ctx context.Context, service string, pubKey []byte, cert []byte, expiry time.Time,
) (err error) {
	res, err := r.db.ExecContext(ctx, queryCreate, service, pubKey, cert, expiry)
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

func (r *CertificateAuthority) Delete(ctx context.Context, service string) error {
	var pubKey, cert []byte

	if err := r.db.QueryRowContext(ctx, queryGet, service).Scan(&pubKey, &cert); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		return err
	}

	res, err := r.db.ExecContext(ctx, queryDelete, service)
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

func (r *CertificateAuthority) Close() error {
	if r.done != nil {
		r.done()
	}

	return r.db.Close()
}

func (r *CertificateAuthority) cleanup(ctx context.Context) error {
	ctx, done := context.WithTimeout(context.Background(), r.cleanupTimeout)
	defer done()

	_, err := r.db.ExecContext(ctx, queryCleanup, time.Now())

	return err
}

func (r *CertificateAuthority) runCron(ctx context.Context) error {
	s, err := schedule.New(
		schedule.WithSchedule(r.cleanupSchedule),
		schedule.WithLogger(r.logger),
		schedule.WithTrace(r.tracer),
		schedule.WithMetrics(r.m),
	)
	if err != nil {
		return err
	}

	exec, err := executor.New("ca_cleanup",
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
