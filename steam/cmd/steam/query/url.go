package query

import (
	"context"
	"net/http"
	"strings"
	"time"
)

const (
	baseURL      = "https://store.steampowered.com/api/appdetails/"
	paramAppIDs  = "appids="
	paramCountry = "cc="
	paramFilter  = "filters="

	defaultTimeout = time.Minute
)

// NewURL creates a URL from the input parameters, fitting the template below
//
// GET https://store.steampowered.com/api/appdetails/?appids={comma_separated_ids}&cc={country}&filters={filters}
func NewURL(ids, country, filter string) string {
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

func NewRequest(ctx context.Context, url string) (*http.Response, error) {
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
