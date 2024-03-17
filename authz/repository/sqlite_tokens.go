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
	minAllocChallenges = 2
	minAllocTokens     = 10

	queryChallengesCreate = `
INSERT INTO challenges (service_id, challenge, expiry)
VALUES ((SELECT id FROM services WHERE name = ?), ?, ?)
`

	queryChallengesList = `
SELECT challenge, expiry FROM challenges
WHERE service_id = (
	SELECT id FROM services WHERE name = ?
)
  AND expiry > ?
ORDER BY expiry DESC
`

	queryChallengesDelete = `
DELETE FROM challenges WHERE service_id = (
	SELECT id FROM services WHERE name = ?
) AND challenge = ?
`

	queryChallengesCleanup = `
DELETE FROM challenges WHERE expiry < ?
`

	queryTokensCreate = `
INSERT INTO tokens (service_id, token, expiry)
VALUES ((SELECT id FROM services WHERE name = ?), ?, ?)
`

	queryTokensList = `
SELECT token, expiry FROM tokens
WHERE service_id = (
	SELECT id FROM services WHERE name = ?
)
  AND expiry > ?
ORDER BY expiry DESC
`

	queryTokensDelete = `
DELETE FROM tokens WHERE service_id = (
	SELECT id FROM services WHERE name = ?
) AND token = ?
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

type Token struct {
	Raw    []byte
	Expiry time.Time
}

type Challenge struct {
	Raw    []byte
	Expiry time.Time
}

func NewToken(db *sql.DB, opts ...cfg.Option[Config]) (*Tokens, error) {
	config := cfg.Set(defaultConfig(), opts...)

	if config.cleanupTimeout <= 0 {
		config.cleanupTimeout = defaultCleanupTimeout
	}

	if config.cleanupSchedule == "" {
		config.cleanupSchedule = defaultCleanupSchedule
	}

	ctx, done := context.WithCancel(context.Background())

	repo := &Tokens{
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

func (r *Tokens) ListChallenges(ctx context.Context, service string) (challenges []Challenge, err error) {
	rows, err := r.db.QueryContext(ctx, queryChallengesList, service)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	challenges = make([]Challenge, 0, minAllocChallenges)

	for rows.Next() {
		var (
			expiryMillis int64
			challenge    = Challenge{}
		)

		if rows.Scan(&challenge.Raw, &expiryMillis); err != nil {
			return nil, err
		}

		challenge.Expiry = time.UnixMilli(expiryMillis)

		challenges = append(challenges, challenge)
	}

	if err = rows.Close(); err != nil {
		return nil, err
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if len(challenges) == 0 {
		return nil, ErrNotFound
	}

	return challenges, nil
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

func (r *Tokens) DeleteChallenge(ctx context.Context, service string, challenge []byte) error {
	res, err := r.db.ExecContext(ctx, queryChallengesDelete, service, challenge)
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

func (r *Tokens) ListTokens(ctx context.Context, service string) (tokens []Token, err error) {
	rows, err := r.db.QueryContext(ctx, queryTokensList, service, time.Now().UnixMilli())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tokens = make([]Token, 0, minAllocTokens)

	for rows.Next() {
		var (
			token Token
			exp   int64
		)

		if err = rows.Scan(&token.Raw, &exp); err != nil {
			return nil, err
		}

		token.Expiry = time.UnixMilli(exp)

		tokens = append(tokens, token)
	}

	if err = rows.Close(); err != nil {
		return nil, err
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if len(tokens) == 0 {
		return nil, ErrNotFound
	}

	return tokens, nil
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

func (r *Tokens) DeleteToken(ctx context.Context, service string, token []byte) error {
	res, err := r.db.ExecContext(ctx, queryTokensDelete, service, token)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}

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
