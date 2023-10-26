package ping

import (
	"context"
	"net/http"
	"time"

	"github.com/zalgonoise/x/cfg"
)

const okStatusLimit = 399

type Checker struct {
	url     string
	timeout time.Duration
}

func NewChecker(options ...cfg.Option[Config]) (*Checker, error) {
	// apply the input options on top of the defined default; the config is a value, not a pointer, in this case.
	config := cfg.Set(defaultConfig, options...)

	if err := validateURL(config); err != nil {
		return nil, err
	}

	// either use the config or pass it along to the data structure if it makes sense that way.
	return &Checker{
		url:     config.url,
		timeout: config.timeout,
	}, nil
}

func (c *Checker) Up(ctx context.Context) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, c.url, http.NoBody)
	if err != nil {
		return false, err
	}

	res, err := (&http.Client{
		Transport: http.DefaultTransport,
		Timeout:   c.timeout,
	}).Do(req)
	if err != nil {
		return false, err
	}

	defer res.Body.Close()

	return res.StatusCode < okStatusLimit, nil
}
