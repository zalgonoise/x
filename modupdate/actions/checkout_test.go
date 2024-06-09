package actions

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/modupdate/config"
)

func TestCheckout(t *testing.T) {
	for _, testcase := range []struct {
		name string
		repo *config.Repository
		cfg  *config.Checkout
		err  error
	}{
		{
			name: "Success/PublicCheckout",
			repo: &config.Repository{
				Path: "github.com/zalgonoise/micron",
			},
			cfg: &config.Checkout{
				Path: "./testdata/micron",
			},
		},
		// private checkout also works
	} {
		t.Run(testcase.name, func(t *testing.T) {
			defer func() {
				require.NoError(t, os.RemoveAll(testcase.cfg.Path))
			}()

			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}))

			a := NewModUpdate(&config.Config{
				Repository: *testcase.repo,
				Checkout:   *testcase.cfg,
			}, logger)

			err := a.Checkout(context.Background())

			require.ErrorIs(t, err, testcase.err)
		})
	}
}
