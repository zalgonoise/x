package service

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zalgonoise/x/torque/database"
	"github.com/zalgonoise/x/torque/repository"
)

func TestLoad(t *testing.T) {
	cleanup(t)

	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	db, err := database.Open("testdata/local.db")
	require.NoError(t, err)

	require.NoError(t, database.Migrate(context.Background(), db, logger))

	repo := repository.NewSQLite(db, logger)

	service := NewService(repo, logger)

	err = service.Load()
	require.NoError(t, err)

	require.NoError(t, db.Close())
}

func cleanup(t *testing.T) {
	err := os.Remove("testdata/local.db")
	if errors.Is(err, os.ErrNotExist) {
		return
	}

	require.NoError(t, err)
}
