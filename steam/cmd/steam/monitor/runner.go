package monitor

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"sync"

	"github.com/zalgonoise/x/steam"
	"github.com/zalgonoise/x/steam/cmd/steam/alert"
	"github.com/zalgonoise/x/steam/cmd/steam/query"
	"github.com/zalgonoise/x/steam/pb/proto/steam/store/v1"
)

const (
	priceFilter = "price_overview"
)

type runner struct {
	logger *slog.Logger
	req    request

	maxInterval int

	mu     *sync.Mutex
	onSale map[string]int
}

type request struct {
	ids            string
	country        string
	platform       string
	url            string
	targetDiscount int
}

func newRunner(logger *slog.Logger, req request, interval int) *runner {
	ids := strings.Split(req.ids, ",")

	return &runner{
		logger:      logger,
		req:         req,
		maxInterval: interval,
		mu:          &sync.Mutex{},
		onSale:      make(map[string]int, len(ids)),
	}
}

func (r *runner) Run(ctx context.Context) error {
	return r.queryPrices(ctx, r.logger, r.req.ids, r.req.country, r.req.platform, r.req.url, r.req.targetDiscount)
}

func (r *runner) queryPrices(
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
	return r.evaluatePrices(ctx, logger, platform, url, targetDiscount, priceOverview)
}

func (r *runner) evaluatePrices(
	ctx context.Context, logger *slog.Logger,
	platform, url string, targetDiscount int,
	priceOverview map[string]*store.PriceOverview,
) error {
	for appID, data := range priceOverview {
		if discount := data.DiscountPercent; int(discount) <= targetDiscount {
			r.mu.Lock()
			onSaleDays, ok := r.onSale[appID]
			r.onSale[appID] = 0
			r.mu.Unlock()

			if ok && onSaleDays > 0 {
				if err := alert.SendMessage(ctx, logger, platform, url, alert.OffSale(appID, data)); err != nil {
					return err
				}
			}

			logger.InfoContext(ctx, "discount isn't low enough",
				slog.String("appID", appID),
				slog.String("final_price", data.GetFinalFormatted()),
				slog.Int("cur_discount_percent", int(data.GetDiscountPercent())),
				slog.Int("target_discount_percent", targetDiscount),
			)

			continue
		}

		r.mu.Lock()
		onSaleDays, ok := r.onSale[appID]
		r.onSale[appID]++
		r.mu.Unlock()

		if ok && onSaleDays > 0 {
			if onSaleDays%r.maxInterval == 0 {
				err := alert.SendMessage(ctx, logger, platform, url, alert.StillOnSale(appID, data))
				if err != nil {
					return err
				}

				continue
			}

			// product still on sale
			slog.InfoContext(ctx, "discount is still up for this product",
				slog.String("appID", appID),
				slog.String("final_price", data.GetFinalFormatted()),
				slog.Int("cur_discount_percent", int(data.GetDiscountPercent())),
			)

			continue
		}

		// exec webhook if current price is lower or equal to discount ratio
		if err := alert.SendMessage(ctx, logger, platform, url, alert.OnSale(appID, data)); err != nil {
			return err
		}
	}

	return nil
}
