package alert

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"strings"

	discordwh "github.com/zalgonoise/x/discord/webhook"
	slackwh "github.com/zalgonoise/x/slack/webhook"

	"github.com/zalgonoise/x/steam"
	"github.com/zalgonoise/x/steam/cmd/steam/query"
	"github.com/zalgonoise/x/steam/pb/proto/steam/store/v1"
)

const (
	priceFilter    = "price_overview"
	productBaseURL = "https://store.steampowered.com/app"
)

var (
	errEmptyID          = errors.New("empty app ID")
	errInvalidPlatform  = errors.New("invalid platform")
	errEmptyURL         = errors.New("empty webhook URL")
	errEmptyTargetPrice = errors.New("empty target price")
)

func Exec(ctx context.Context, logger *slog.Logger, args []string) (error, int) {
	fs := flag.NewFlagSet("alert", flag.ExitOnError)

	ids := fs.String("ids", "", "comma-separated list of app ID values")
	country := fs.String("country", "", "country code (2-character-long)")
	platform := fs.String("platform", "logger", "target platform where to post (logger; slack; discord)")
	url := fs.String("url", "", "webhook target URL (platform: slack; discord)")
	targetDiscount := fs.Int("target_discount", 50, "target discount percent")

	if err := fs.Parse(args); err != nil {
		return err, 1
	}

	if *ids == "" {
		return errEmptyID, 1
	}

	if *targetDiscount == 0 {
		return errEmptyTargetPrice, 1
	}

	if *platform == "" {
		*platform = "logger"
	}

	if *url == "" && *platform != "logger" {
		return errEmptyURL, 1
	}

	if err := QueryPrices(ctx, logger, *ids, *country, *platform, *url, *targetDiscount); err != nil {
		return err, 1
	}

	return nil, 0
}

func QueryPrices(
	ctx context.Context,
	logger *slog.Logger,
	ids, country, platform, url string,
	targetDiscount int,
) error {
	storeURL := query.NewURL(ids, country, priceFilter)

	res, err := query.NewRequest(ctx, storeURL)

	defer res.Body.Close()

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	priceOverview, err := steam.GetPriceOverview(buf)
	if err != nil {
		return err
	}

	// eval against target
	return EvaluatePrices(ctx, logger, platform, url, targetDiscount, priceOverview)
}

func EvaluatePrices(
	ctx context.Context, logger *slog.Logger,
	platform, url string, targetDiscount int,
	priceOverview map[string]*store.PriceOverview,
) error {
	for appID, data := range priceOverview {
		if discount := data.DiscountPercent; int(discount) <= targetDiscount {
			logger.InfoContext(ctx, "discount isn't low enough",
				slog.String("appID", appID),
				slog.String("final_price", data.GetFinalFormatted()),
				slog.Int("cur_discount_percent", int(data.GetDiscountPercent())),
				slog.Int("target_discount_percent", targetDiscount),
			)

			continue
		}

		// exec webhook if current price is lower or equal to discount ratio
		if err := SendMessage(ctx, logger, platform, url, OnSale(appID, data)); err != nil {
			return err
		}
	}

	return nil
}

func SendMessage(
	ctx context.Context, logger *slog.Logger,
	platform, url, payload string,
) error {
	switch strings.ToLower(platform) {
	case "logger":
		logger.InfoContext(ctx, payload)

		return nil
	case "slack":
		w, err := slackwh.New(url, slackwh.WithLogger(logger))
		if err != nil {
			return err
		}

		if _, err = w.Execute(ctx, payload); err != nil {
			return err
		}

		return nil
	case "discord":
		w, err := discordwh.New(url, discordwh.WithLogger(logger))
		if err != nil {
			return err
		}
		if _, err = w.Execute(ctx, payload); err != nil {
			return err
		}

		return nil
	default:
		return fmt.Errorf("%w: %s", errInvalidPlatform, platform)
	}
}

func OnSale(appID string, data *store.PriceOverview) string {
	return fmt.Sprintf(
		`Woah! The app %s is currently with a %d percent discount!!

	Initial price: %s
	Current price: %s

Check it out at: %s/%s`,
		appID,
		data.GetDiscountPercent(),
		data.GetInitialFormatted(),
		data.GetFinalFormatted(),
		productBaseURL,
		appID,
	)
}

func StillOnSale(appID string, data *store.PriceOverview) string {
	return fmt.Sprintf(
		`What!? The app %s is still with a %d percent discount!!

	Initial price: %s
	Current price: %s

Check it out at: %s/%s`,
		appID,
		data.GetDiscountPercent(),
		data.GetInitialFormatted(),
		data.GetFinalFormatted(),
		productBaseURL,
		appID,
	)
}

func OffSale(appID string, data *store.PriceOverview) string {
	return fmt.Sprintf(
		`Shoot! The app %s is no longer on sale... Better luck next time!

	Current discount: %d%
	Current price: %s

Check it out at: %s/%s`,
		appID,
		data.GetDiscountPercent(),
		data.GetFinalFormatted(),
		productBaseURL,
		appID,
	)
}
