package store

import (
	"context"
	"errors"
	"flag"
	"io"
	"log/slog"

	"github.com/zalgonoise/x/steam"
	"github.com/zalgonoise/x/steam/cmd/steam/filters"
	"github.com/zalgonoise/x/steam/cmd/steam/query"
)

var (
	errEmptyID = errors.New("empty app ID")
)

func Exec(ctx context.Context, logger *slog.Logger, args []string) (error, int) {
	fs := flag.NewFlagSet("store", flag.ExitOnError)

	ids := fs.String("ids", "", "comma-separated list of app ID values")
	country := fs.String("country", "", "country code (2-character-long)")
	filter := fs.String("filter", "", "object query filter")

	fs.Parse(args)

	if *ids == "" {
		return errEmptyID, 1
	}

	url := query.NewURL(*ids, *country, *filter)

	res, err := query.NewRequest(ctx, url)

	defer res.Body.Close()

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return err, 1
	}

	if p, ok := filters.ValidPrinters[*filter]; ok {
		p(ctx, logger, buf)

		return nil, 0
	}

	data, err := steam.UnmarshalJSON(buf)
	if err != nil {
		return err, 1
	}

	details := data.GetAppDetails()
	logger.InfoContext(ctx, "received app details",
		slog.Int("num_results", len(details)),
	)

	for appID, appDetails := range details {
		logger.InfoContext(ctx, "describing app listing",
			slog.Bool("status", appDetails.GetSuccess()),
			slog.String("app_id", appID),
			slog.String("name", appDetails.GetData().GetName()),
			slog.String("current_price", appDetails.GetData().GetPriceOverview().GetFinalFormatted()),
		)
	}

	return nil, 0
}
