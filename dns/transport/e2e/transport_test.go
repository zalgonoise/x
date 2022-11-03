// Package e2e will expect a running instance of the app, where this
// test file will hit the supported endpoints to ensure they are working
// as intended
//
// to run the tests, run the app via Docker or
// `sudo go run examples/core_memmap/main.go`
package e2e

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
)

func httpGet(endpoint string) ([]byte, int, error) {
	res, err := http.Get(fmt.Sprintf("http://localhost:%v%s", 8080, endpoint))
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, 0, err
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, 0, err
	}
	return b, res.StatusCode, nil
}

func TestTransport(t *testing.T) {
	t.Run("HTTP", func(t *testing.T) {
		t.Run("StartDNS", func(t *testing.T) {
			b, status, err := httpGet("/dns/start")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if status != 200 {
				t.Errorf("unexpected status code: wanted %v ; got %v -- body: %s", 200, status, string(b))
			}
		})
		t.Run("ReloadDNS", func(t *testing.T) {
			b, status, err := httpGet("/dns/reload")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if status != 200 {
				t.Errorf("unexpected status code: wanted %v ; got %v -- body: %s", 200, status, string(b))
			}
		})
		t.Run("StopDNS", func(t *testing.T) {
			b, status, err := httpGet("/dns/stop")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if status != 200 {
				t.Errorf("unexpected status code: wanted %v ; got %v -- body: %s", 200, status, string(b))
			}
		})
	})
}
