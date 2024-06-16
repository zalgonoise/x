package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zalgonoise/x/modupdate/config"
	"github.com/zalgonoise/x/modupdate/database"
	"github.com/zalgonoise/x/modupdate/log"
)

func TestRepository_AddTask_ListTasks(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		input *config.Task
	}{
		{
			name: "Success_WithTask",
			input: &config.Task{
				Repository: config.Repository{
					Path:       "github.com/zalgonoise/x",
					ModulePath: "audio",
					Branch:     "master",
				},
				Checkout: config.Checkout{
					Persist: true,
					Path:    "local/src/github.com/zalgonoise/x",
				},
				Update: config.Update{},
				Check: config.Check{
					Skip: true,
				},
				Push: config.Push{
					DryRun:        true,
					CommitMessage: "audio: go.mod: updated dependencies",
				},
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			logger := log.New("debug")

			db, err := database.OpenSQLite("", database.ReadWritePragmas(), logger)
			require.NoError(t, err)

			ctx := context.Background()

			require.NoError(t, database.MigrateSQLite(ctx, db, logger))

			repo := NewRepository(db)

			require.NoError(t, repo.AddTask(ctx, testcase.input))

			task := &taskID{}

			row := db.QueryRowContext(ctx, queryRepositoryID+filterByURIModuleAndBranch,
				testcase.input.Repository.Path, testcase.input.Repository.ModulePath, testcase.input.Repository.Branch,
			)

			require.NoError(t, row.Scan(&task.id))
			require.NoError(t, row.Err())

			require.NotEmpty(t, task.id)

			tasks, err := repo.ListTasks(ctx)
			require.NoError(t, err)

			require.Len(t, tasks, 1)
			t.Log(tasks[0])
		})
	}
}
