package store

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/zalgonoise/x/steam"
)

const (
	baseURL        = "https://store.steampowered.com/api/appdetails/"
	paramAppIDs    = "appids="
	paramCountry   = "cc="
	paramFilter    = "filters="
	defaultTimeout = time.Minute
)

var (
	errEmptyID = errors.New("empty app ID")
)

// newURL creates a URL from the input parameters, fitting the template below
//
// GET https://store.steampowered.com/api/appdetails/?appids={comma_separated_ids}&cc={country}&filters={filters}
func newURL(ids, country, filter string) string {
	sb := &strings.Builder{}

	sb.WriteString(baseURL)
	sb.WriteByte('?')
	sb.WriteString(paramAppIDs)
	sb.WriteString(ids)

	if country != "" {
		sb.WriteByte('&')
		sb.WriteString(paramCountry)
		sb.WriteString(country)
	}

	if filter != "" {
		sb.WriteByte('&')
		sb.WriteString(paramFilter)
		sb.WriteString(filter)
	}

	return sb.String()
}

func newReq(ctx context.Context, url string) (*http.Response, error) {
	client := &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   defaultTimeout,
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func Exec(ctx context.Context, logger *slog.Logger, args []string) (error, int) {
	fs := flag.NewFlagSet("store", flag.ExitOnError)

	ids := fs.String("ids", "", "comma-separated list of app ID values")
	country := fs.String("country", "", "country code (2-character-long)")
	filter := fs.String("filter", "", "object query filter")

	fs.Parse(args)

	if *ids == "" {
		return errEmptyID, 1
	}

	url := newURL(*ids, *country, *filter)

	res, err := newReq(ctx, url)

	defer res.Body.Close()

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return err, 1
	}

	if fn, ok := validFilters[*filter]; ok {
		if err = fn(ctx, logger, buf); err != nil {
			return err, 1
		}

		return nil, 0
	}

	b := bytes.NewBuffer(nil)
	json.Indent(b, buf, "", "  ")
	fmt.Println(b.String())

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
