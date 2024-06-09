package actions

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zalgonoise/x/modupdate/config"
)

func TestPush(t *testing.T) {
	for _, testcase := range []struct {
		name     string
		repo     config.Repository
		checkout config.Checkout
		update   config.Update
		push     config.Push
		err      error
	}{
		{
			name: "Success/PublicCheckoutUpdateAndPush",
			repo: config.Repository{
				Path:   "github.com/zalgonoise/micron",
				Branch: "master",
			},
			checkout: config.Checkout{
				Path: "./testdata/micron",
			},
			update: config.Update{},
			push: config.Push{
				DryRun: true,
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			defer func() {
				require.NoError(t, os.RemoveAll(testcase.checkout.Path))
			}()

			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}))

			a := NewModUpdate(&config.Config{
				Repository: testcase.repo,
				Checkout:   testcase.checkout,
				Update:     testcase.update,
				Push:       testcase.push,
			}, logger)

			ctx := context.Background()

			require.NoError(t, a.Checkout(ctx))
			require.NoError(t, a.Update(ctx))
			require.ErrorIs(t, a.Push(ctx), testcase.err)
		})
	}
}
