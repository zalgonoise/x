package database

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const minBlockSize = 32_000

type migration struct {
	table  string
	create string
}

func MigrateSQLite(ctx context.Context, db *sql.DB, dir string, logger *slog.Logger) error {
	start := time.Now()

	if err := runMigrations(ctx, db, newSchema()...); err != nil {
		return err
	}

	data, err := readDataDir(ctx, dir, logger)
	if err != nil {
		return err
	}

	if err := insertData(ctx, db, data, minBlockSize, logger); err != nil {
		return err
	}

	logger.InfoContext(ctx, "operation completed", slog.Duration("time_elapsed", time.Since(start)))

	return nil
}

func runMigrations(ctx context.Context, db *sql.DB, migrations ...migration) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	for i := range migrations {
		r, err := tx.QueryContext(ctx, fmt.Sprintf(checkTableExists, migrations[i].table))
		if err != nil {
			return err
		}

		var count int

		if !r.Next() {
			return r.Err()
		}

		if err = r.Scan(&count); err != nil {
			_ = r.Close()

			return err
		}

		_ = r.Close()

		if count == 1 {
			continue
		}

		_, err = tx.ExecContext(ctx, migrations[i].create)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func readDataDir(ctx context.Context, dir string, logger *slog.Logger) ([]int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	logger.InfoContext(ctx, "scanned directory", slog.Int("num_files", len(entries)))

	total := make([]int, 0, len(entries)*maxAlloc)
	errs := make([]error, 0, len(entries))

	for i := range entries {
		logger.InfoContext(ctx, "extracting primes from file", slog.String("filename", entries[i].Name()))

		data, extractErr := extract(ctx, logger, path.Join(dir, entries[i].Name()))
		if extractErr != nil {
			errs = append(errs, err)

			continue
		}

		total = append(total, data...)
	}

	logger.InfoContext(ctx, "extracted primes from input file(s)",
		slog.Int("num_primes", len(total)),
		slog.Int("num_errors", len(errs)),
	)

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return total, nil
}

func extract(ctx context.Context, logger *slog.Logger, path string) ([]int, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			logger.WarnContext(ctx, "error closing file",
				slog.String("error", err.Error()),
				slog.String("filename", file.Name()),
			)
		}
	}()

	scanner := bufio.NewScanner(file)

	values := make([]int, 0, maxAlloc)
	errs := make([]error, 0, maxAlloc)

	for scanner.Scan() {
		value, convErr := strconv.Atoi(scanner.Text())
		if convErr != nil {
			errs = append(errs, convErr)

			continue
		}

		values = append(values, value)
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return values, nil
}

func insertData(ctx context.Context, db *sql.DB, data []int, blockSize int, logger *slog.Logger) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	logger.InfoContext(ctx, "preparing transaction", slog.Int("num_primes", len(data)))

	var offset int

	for i := 0; i < len(data); i += blockSize {
		to := i + blockSize - 1
		if to > len(data) {
			to = len(data)
		}

		query, args := buildStatement(data[i:to])
		if _, err = tx.ExecContext(ctx, query, args...); err != nil {
			return err
		}

		offset++

		if offset >= 10 {
			offset = 0

			logger.InfoContext(ctx, "executing insert query", slog.Int("cur_index", i))
		}
	}

	logger.InfoContext(ctx, "committing transaction")

	return tx.Commit()
}

func buildStatement(data []int) (string, []any) {
	sb := &strings.Builder{}

	sb.WriteString(`INSERT INTO primes (prime) VALUES `)
	for i := 0; i < len(data); i++ {
		sb.WriteString("(?)")

		if i < len(data)-1 {
			sb.WriteByte(',')
		}
	}

	sb.WriteString(`;`)

	args := make([]any, len(data))
	for i := range data {
		args[i] = data[i]
	}

	return sb.String(), args
}
