package actions

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zalgonoise/x/modupdate/config"
)

func TestUpdate(t *testing.T) {
	for _, testcase := range []struct {
		name     string
		repo     config.Repository
		checkout config.Checkout
		cfg      config.Update
		err      error
	}{
		{
			name: "Success/PublicCheckoutAndUpdate",
			repo: config.Repository{
				Path: "github.com/zalgonoise/micron",
			},
			checkout: config.Checkout{
				Path: "./testdata/micron",
			},
			cfg: config.Update{},
		},
		{
			name: "Success/PublicCheckoutAndUpdate/CustomGoBin",
			repo: config.Repository{
				Path: "github.com/zalgonoise/micron",
			},
			checkout: config.Checkout{
				Path: "./testdata/micron",
			},
			cfg: config.Update{
				GoBin: "~/go/go1.22.1/bin/go",
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
				Update:     testcase.cfg,
			}, logger)

			ctx := context.Background()

			require.NoError(t, a.Checkout(ctx))
			require.ErrorIs(t, a.Update(ctx), testcase.err)
		})
	}
}
