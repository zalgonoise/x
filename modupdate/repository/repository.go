package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/zalgonoise/x/iter"
	"github.com/zalgonoise/x/modupdate/config"
)

const (
	minAlloc = 64

	typeCheckout  = "checkout"
	typeUpdateGit = "update/git"
	typeUpdateGo  = "update/go"
	typePush      = "push"
	typePushFiles = "push/files"

	queryRepositoryAndTasks = `
SELECT id, uri, module, branch, username, token, cron_schedule, dry_run, fs_path, commit_message 
	FROM repositories`

	queryRepositoryID = `
SELECT id FROM repositories`

	queryOverrides = `
SELECT command FROM overrides
	WHERE id = ?
	AND type = ?`

	filterByURIModuleAndBranch = `
		WHERE uri = ?
		AND module = ?
		AND branch = ?;`

	deleteRepo = `
DELETE FROM repositories WHERE id = ?`

	deleteOverrides = `
DELETE FROM overrides WHERE id = ?`

	insertRepositoryAndTasks = `
INSERT INTO repositories (
	id, uri, module, branch, username, token, cron_schedule, dry_run, fs_path, commit_message
) VALUES (
	?, ?, ?, ?, ?, ?, ?, ?, ?, ?
)`

	insertOverrides = `
INSERT INTO overrides (
	id, type, command
) VALUES (
	?, ?, ?,
)`
)

var ErrSeqFailed = errors.New("processing task sequence failed")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type taskID struct {
	id sql.Null[string]
}

func (t *taskID) Scan(row *sql.Rows) error {
	return row.Scan(
		&t.id,
	)
}

type task struct {
	id            sql.Null[string]
	uri           sql.Null[string]
	module        sql.Null[string]
	branch        sql.Null[string]
	username      sql.Null[string]
	token         sql.Null[string]
	cronSchedule  sql.Null[string]
	dryRun        sql.Null[int]
	fsPath        sql.Null[string]
	commitMessage sql.Null[string]
}

func (t *task) Scan(row *sql.Rows) error {
	return row.Scan(
		&t.id,
		&t.uri,
		&t.module,
		&t.branch,
		&t.username,
		&t.token,
		&t.cronSchedule,
		&t.dryRun,
		&t.fsPath,
		&t.commitMessage,
	)
}

type override string

func (o *override) Scan(row *sql.Rows) error {
	return row.Scan(&o)
}

func (r *Repository) ListTasks(ctx context.Context) ([]*config.Task, error) {
	seq, err := iter.QueryContext[*task](ctx, r.db, queryRepositoryAndTasks)
	if errors.Is(err, sql.ErrNoRows) {
		return []*config.Task{}, nil
	}

	if err != nil {
		return nil, err
	}

	configs := make([]*config.Task, 0, minAlloc)
	errs := make([]error, 0, minAlloc)

	seq(r.collectSeq(ctx, &configs, &errs))

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return configs, nil
}

func (r *Repository) collectSeq(ctx context.Context, configs *[]*config.Task, errs *[]error) func(t *task, err error) bool {
	return func(t *task, err error) bool {
		if err != nil {
			*errs = append(*errs, err)

			// collect all valid tasks
			return true
		}
		dryRun := false

		if t.dryRun.V == 1 {
			dryRun = true
		}

		c := &config.Task{
			CronSchedule: t.cronSchedule.V,
			Repository: config.Repository{
				Path:       t.uri.V,
				ModulePath: t.module.V,
				Branch:     t.branch.V,
				Username:   t.username.V,
				Token:      t.token.V,
			},
			Checkout: config.Checkout{
				Persist: true,
				Path:    t.fsPath.V,
			},
			Push: config.Push{
				DryRun:        dryRun,
				CommitMessage: t.commitMessage.V,
			},
		}

		overrides, err := addOverrides(ctx, r.db, t.id.V, typeCheckout)
		if err != nil {
			*errs = append(*errs, err)

			// collect all valid tasks
			return true
		}

		c.Checkout.CommandOverrides = overrides

		overrides, err = addOverrides(ctx, r.db, t.id.V, typeUpdateGit)
		if err != nil {
			*errs = append(*errs, err)

			// collect all valid tasks
			return true
		}

		c.Update.GitCommandOverrides = overrides

		overrides, err = addOverrides(ctx, r.db, t.id.V, typeUpdateGo)
		if err != nil {
			*errs = append(*errs, err)

			// collect all valid tasks
			return true
		}

		c.Update.GoCommandOverrides = overrides

		overrides, err = addOverrides(ctx, r.db, t.id.V, typePush)
		if err != nil {
			*errs = append(*errs, err)

			// collect all valid tasks
			return true
		}

		c.Push.CommandOverrides = overrides

		overrides, err = addOverrides(ctx, r.db, t.id.V, typePushFiles)
		if err != nil {
			*errs = append(*errs, err)

			// collect all valid tasks
			return true
		}

		c.Push.FilesOverride = overrides

		*configs = append(*configs, c)

		return true
	}
}

func (r *Repository) Close() error {
	return r.db.Close()
}

func addOverrides(ctx context.Context, tx transactioner, id, typ string) ([]string, error) {
	seq, err := iter.QueryContext[*override](ctx, tx, queryOverrides, id, typ)
	if errors.Is(err, sql.ErrNoRows) {
		return []string{}, nil
	}

	if err != nil {
		return nil, err
	}

	overrides := make([]string, 0, minAlloc)

	if !seq(func(o *override, e error) bool {
		if e != nil {
			err = e

			return false
		}

		overrides = append(overrides, string(*o))

		return true
	}) {
		return nil, err
	}

	return overrides, nil
}

func (r *Repository) AddTask(ctx context.Context, cfg *config.Task) error {
	if err := r.DeleteTask(ctx, cfg.Repository.Path, cfg.Repository.ModulePath, cfg.Repository.Branch); err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	id, err := createTask(ctx, tx, cfg)
	if err != nil {
		return err
	}

	if len(cfg.Checkout.CommandOverrides) > 0 {
		for i := range cfg.Checkout.CommandOverrides {
			if err := createOverrides(ctx, tx, id, typeCheckout, cfg.Checkout.CommandOverrides[i]); err != nil {
				return err
			}
		}
	}

	if len(cfg.Update.GitCommandOverrides) > 0 {
		for i := range cfg.Update.GitCommandOverrides {
			if err := createOverrides(ctx, tx, id, typeUpdateGit, cfg.Update.GitCommandOverrides[i]); err != nil {
				return err
			}
		}
	}

	if len(cfg.Update.GoCommandOverrides) > 0 {
		for i := range cfg.Update.GoCommandOverrides {
			if err := createOverrides(ctx, tx, id, typeUpdateGo, cfg.Update.GoCommandOverrides[i]); err != nil {
				return err
			}
		}
	}

	if len(cfg.Push.CommandOverrides) > 0 {
		for i := range cfg.Push.CommandOverrides {
			if err := createOverrides(ctx, tx, id, typePush, cfg.Push.CommandOverrides[i]); err != nil {
				return err
			}
		}
	}

	if len(cfg.Push.FilesOverride) > 0 {
		for i := range cfg.Push.FilesOverride {
			if err := createOverrides(ctx, tx, id, typePushFiles, cfg.Push.FilesOverride[i]); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *Repository) DeleteTask(ctx context.Context, uri, module, branch string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	seq, err := iter.QueryContext[*taskID](ctx, r.db, queryRepositoryID+filterByURIModuleAndBranch,
		uri, module, branch)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}

	if err != nil {
		return err
	}

	errs := make([]error, 0, minAlloc)
	ids := make([]string, 0, minAlloc)

	if !seq(func(t *taskID, err error) bool {
		if err != nil {
			errs = append(errs, err)

			return false
		}

		// nothing to remove
		if !t.id.Valid {
			return true
		}

		ids = append(ids, t.id.V)

		if err := removeTask(ctx, tx, t.id.V); err != nil {
			errs = append(errs, err)

			return false
		}

		return true
	}) {
		errs = append(errs, ErrSeqFailed, tx.Rollback())

		return errors.Join(errs...)
	}

	return tx.Commit()
}

type transactioner interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func createTask(ctx context.Context, tx transactioner, cfg *config.Task) (string, error) {
	id := uuid.New().String()

	dryRun := 0
	if cfg.Push.DryRun {
		dryRun = 1
	}

	res, err := tx.ExecContext(ctx, insertRepositoryAndTasks,
		id, cfg.Repository.Path, cfg.Repository.ModulePath, cfg.Repository.Branch,
		cfg.Repository.Username, cfg.Repository.Token,
		cfg.CronSchedule, dryRun, cfg.Checkout.Path, cfg.Push.CommitMessage,
	)
	if err != nil {
		return "", err
	}

	if _, err = res.RowsAffected(); err != nil {
		return "", err
	}

	return id, nil
}

func createOverrides(ctx context.Context, tx transactioner, id, typ, command string) error {
	res, err := tx.ExecContext(ctx, insertOverrides, id, typ, command)
	if err != nil {
		return err
	}

	if _, err = res.RowsAffected(); err != nil {
		return err
	}

	return nil
}

func removeTask(ctx context.Context, tx transactioner, id string) error {
	res, err := tx.ExecContext(ctx, deleteOverrides, id)
	if err != nil {
		return err
	}

	if _, err = res.RowsAffected(); err != nil {
		return err
	}

	res, err = tx.ExecContext(ctx, deleteRepo, id)
	if err != nil {
		return err
	}

	if _, err = res.RowsAffected(); err != nil {
		return err
	}

	return nil
}
