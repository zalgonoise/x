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
	errEmptyPlatform    = errors.New("empty platform")
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
		if discount := data.DiscountPercent; int(discount) <= *targetDiscount {
			continue
		}

		// exec webhook if current price is lower or equal to discount ratio
		if err = sendMessage(ctx, logger, *platform, *url, appID, data); err != nil {
			return err, 1
		}
	}

	return nil, 0
}

func sendMessage(
	ctx context.Context, logger *slog.Logger,
	platform, url, appID string, data *store.PriceOverview,
) error {
	switch strings.ToLower(platform) {
	case "logger":
		logger.InfoContext(ctx, textAlert(appID, data))

		return nil
	case "slack":
		w, err := slackwh.New(url, slackwh.WithLogger(logger))
		if err != nil {
			return err
		}

		if _, err = w.Execute(ctx, textAlert(appID, data)); err != nil {
			return err
		}

		return nil
	case "discord":
		w, err := discordwh.New(url, discordwh.WithLogger(logger))
		if err != nil {
			return err
		}
		if _, err = w.Execute(ctx, textAlert(appID, data)); err != nil {
			return err
		}

		return nil
	default:
		return fmt.Errorf("%w: %s", errInvalidPlatform, platform)
	}
}

func textAlert(appID string, data *store.PriceOverview) string {
	return fmt.Sprintf(
		`Woah! The app %s is currently with a %d percent discount!!

	Initial price: %s
	Current price: %s

Check it out at: %s/%s`,
		appID,
		data.DiscountPercent,
		data.InitialFormatted,
		data.FinalFormatted,
		productBaseURL,
		appID,
	)
}
