package alert

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"

	"github.com/zalgonoise/x/steam"
	"github.com/zalgonoise/x/steam/cmd/steam/query"
)

const (
	priceFilter = "price_overview"
)

var (
	errEmptyID          = errors.New("empty app ID")
	errEmptyPlatform    = errors.New("empty platform")
	errEmptyURL         = errors.New("empty webhook URL")
	errEmptyTargetPrice = errors.New("empty target price")
)

func Exec(ctx context.Context, logger *slog.Logger, args []string) (error, int) {
	fs := flag.NewFlagSet("store", flag.ExitOnError)

	ids := fs.String("ids", "", "comma-separated list of app ID values")
	country := fs.String("country", "", "country code (2-character-long)")
	platform := fs.String("platform", "", "target platform where to post")
	url := fs.String("url", "", "webhook target URL")
	target := fs.Int("target_discount", 50, "target discount percent")

	if err := fs.Parse(args); err != nil {
		return err, 1
	}

	if *ids == "" {
		return errEmptyID, 1
	}

	if *target == 0 {
		return errEmptyTargetPrice, 1
	}

	if *platform == "" {
		return errEmptyPlatform, 1
	}

	if *url == "" {
		return errEmptyURL, 1
	}

	// get appdetail with filter
	storeURL := query.NewURL(*ids, *country, priceFilter)

	res, err := query.NewRequest(ctx, storeURL)

	defer res.Body.Close()

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return err, 1
	}

	priceOverview, err := steam.GetPriceOverview(buf)
	if err != nil {
		return err, 1
	}

	fmt.Println(string(buf), priceOverview)

	// eval against target
	for appID, data := range priceOverview {
		if discount := data.DiscountPercent; int(discount) >= *target {
			continue
		}

		// exec webhook if current price is lower or equal to discount ratio
		_ = appID
		_ = data
	}

	return nil, 0
}
