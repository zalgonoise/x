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
	pb "github.com/zalgonoise/x/authz/pb/authz/v1"
	"go.opentelemetry.io/otel/trace"
)

const (
	minAllocCertificates = 2

	queryServicesGet = `
SELECT pub_key FROM services 
WHERE name = ?
`

	queryServicesCreate = `
INSERT INTO services (name, pub_key)
VALUES (?, ?)
`

	queryCertificatesList = `
SELECT cert, expiry FROM certificates 
WHERE service_id = (SELECT id FROM services WHERE name = ?)
AND expiry > ?
ORDER BY expiry DESC
`

	queryCertificatesCreate = `
INSERT INTO certificates (service_id, cert, expiry)
VALUES ((SELECT id FROM services WHERE name = ?),  ?, ?)
`

	queryCertificatesDeleteAll = `
DELETE FROM certificates WHERE service_id = (SELECT id FROM services WHERE name = ?)
`

	queryCertificatesDelete = `
DELETE FROM certificates 
	WHERE service_id = (SELECT id FROM services WHERE name = ?)
	AND cert = ?
`

	queryServicesDelete = `
DELETE FROM services WHERE name = ?
`

	queryServicesCleanup = `
DELETE FROM certificates WHERE expiry < ?
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

type Services struct {
	cleanupTimeout  time.Duration
	cleanupSchedule string

	db *sql.DB

	done context.CancelFunc

	logger *slog.Logger
	m      Metrics
	tracer trace.Tracer
}

type Certificate struct {
	Raw    []byte
	Expiry time.Time
}

func NewService(db *sql.DB, opts ...cfg.Option[Config]) (*Services, error) {
	config := cfg.Set(defaultConfig(), opts...)

	if config.cleanupTimeout <= 0 {
		config.cleanupTimeout = defaultCleanupTimeout
	}

	if config.cleanupSchedule == "" {
		config.cleanupSchedule = defaultCleanupSchedule
	}

	ctx, done := context.WithCancel(context.Background())

	repo := &Services{
		cleanupTimeout:  config.cleanupTimeout,
		cleanupSchedule: config.cleanupSchedule,
		db:              db,
		done:            done,
		logger:          config.logger,
		m:               config.m,
		tracer:          config.tracer,
	}

	if err := repo.runCron(ctx); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *Services) GetService(ctx context.Context, service string) (pubKey []byte, err error) {
	if err = r.db.QueryRowContext(
		ctx, queryServicesGet, service, time.Now().UnixMilli(),
	).Scan(&pubKey); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return pubKey, nil
}

func (r *Services) CreateService(ctx context.Context, service string, pubKey []byte) (err error) {
	res, err := r.db.ExecContext(ctx, queryServicesCreate, service, pubKey)
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

func (r *Services) DeleteService(ctx context.Context, service string) error {
	var pubKey []byte

	if err := r.db.QueryRowContext(ctx, queryServicesGet, service).Scan(&pubKey); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		return err
	}

	_, err := r.db.ExecContext(ctx, queryCertificatesDeleteAll, service)
	if err != nil {
		return err
	}

	res, err := r.db.ExecContext(ctx, queryServicesDelete, service)
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

func (r *Services) ListCertificates(ctx context.Context, service string) (certs []*pb.CertificateResponse, err error) {
	rows, err := r.db.QueryContext(ctx, queryCertificatesList, service, time.Now().UnixMilli())
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	certs = make([]*pb.CertificateResponse, 0, minAllocCertificates)

	for rows.Next() {
		cert := &pb.CertificateResponse{}

		if err = rows.Scan(&cert.Certificate, &cert.ExpiresOn); err != nil {
			return nil, err
		}

		certs = append(certs, cert)
	}

	if err = rows.Close(); err != nil {
		return nil, err
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return certs, nil
}

func (r *Services) CreateCertificate(ctx context.Context, service string, cert []byte, expiry time.Time) error {
	res, err := r.db.ExecContext(ctx, queryCertificatesCreate, service, cert, expiry.UnixMilli())
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

func (r *Services) DeleteCertificate(ctx context.Context, service string, cert []byte) error {
	res, err := r.db.ExecContext(ctx, queryCertificatesDelete, service, cert)
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

func (r *Services) Close() error {
	if r.done != nil {
		r.done()
	}

	return r.db.Close()
}

func (r *Services) cleanup(ctx context.Context) error {
	ctx, done := context.WithTimeout(context.Background(), r.cleanupTimeout)
	defer done()

	_, err := r.db.ExecContext(ctx, queryServicesCleanup, time.Now().UnixMilli())

	return err
}

func (r *Services) runCron(ctx context.Context) error {
	s, err := schedule.New(
		schedule.WithSchedule(r.cleanupSchedule),
		schedule.WithLogger(r.logger),
		schedule.WithTrace(r.tracer),
		schedule.WithMetrics(r.m),
	)
	if err != nil {
		return err
	}

	exec, err := executor.New("services_cleanup",
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
