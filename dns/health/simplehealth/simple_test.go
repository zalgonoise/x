package simplehealth

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/zalgonoise/x/dns/health"
	"github.com/zalgonoise/x/dns/store"
)

var (
	mock200 http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}
	mock500 http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}
)

func httpServer(fn http.HandlerFunc) *http.Server {
	mux := http.NewServeMux()
	s := &http.Server{
		Addr:    fmt.Sprintf(":%v", 48052),
		Handler: mux,
	}
	mux.HandleFunc("/records", fn)
	return s
}

func TestStore(t *testing.T) {
	t.Run("Healthy", func(t *testing.T) {
		wants := &health.StoreReport{
			Status:   health.Healthy,
			Len:      2,
			Duration: 2,
		}

		h := New()

		report := h.Store(2, 2*time.Millisecond)

		if !reflect.DeepEqual(wants, report) {
			t.Errorf("output mismatch error: wanted %v ; got %v", wants, report)
			return
		}
	})
	t.Run("Stopped", func(t *testing.T) {
		wants := &health.StoreReport{
			Status:   health.Stopped,
			Len:      0,
			Duration: 0,
		}

		h := New()

		report := h.Store(0, 0)

		if !reflect.DeepEqual(wants, report) {
			t.Errorf("output mismatch error: wanted %v ; got %v", wants, report)
			return
		}
	})
	t.Run("Running", func(t *testing.T) {
		wants := &health.StoreReport{
			Status:   health.Running,
			Len:      0,
			Duration: 2,
		}

		h := New()

		report := h.Store(0, 2*time.Millisecond)

		if !reflect.DeepEqual(wants, report) {
			t.Errorf("output mismatch error: wanted %v ; got %v", wants, report)
			return
		}
	})
}

func TestDNS(t *testing.T) {
	t.Run("Healthy", func(t *testing.T) {
		wants := &health.DNSReport{
			Enabled: true,
			Status:  health.Healthy,
		}
		h := New()

		report := h.DNS("1.1.1.1:53", "8.8.8.8:53", store.New().Type(store.TypeA.String()).Name("google.com").Build())

		if report.LocalQuery == 0 || report.ExternalQuery == 0 {
			t.Errorf("DNS queries must take time to complete, and cannot be zero")
			return
		}
		report.LocalQuery = 0
		report.ExternalQuery = 0

		if !reflect.DeepEqual(wants, report) {
			t.Errorf("output mismatch error: wanted %v ; got %v", wants, report)
		}
	})
	t.Run("Unhealthy", func(t *testing.T) {
		wants := &health.DNSReport{
			Enabled: false,
			Status:  health.Unhealthy,
		}
		h := New()

		report := h.DNS("1.1.1.1", "8.8.8.8:53", store.New().Type(store.TypeA.String()).Name("google.com").Build())

		if report.LocalQuery == 0 || report.ExternalQuery == 0 {
			t.Errorf("DNS queries must take time to complete, and cannot be zero")
			return
		}
		report.LocalQuery = 0
		report.ExternalQuery = 0

		if !reflect.DeepEqual(wants, report) {
			t.Errorf("output mismatch error: wanted %v ; got %v", wants, report)
		}
	})
	t.Run("Stopped", func(t *testing.T) {
		wants := &health.DNSReport{
			Enabled: false,
			Status:  health.Stopped,
		}
		h := New()

		report := h.DNS("1.1.1.1:53", "", nil)

		if !reflect.DeepEqual(wants, report) {
			t.Errorf("output mismatch error: wanted %v ; got %v", wants, report)
		}
	})
}

func TestHTTP(t *testing.T) {
	s200 := httpServer(mock200)
	go func() {
		err := s200.ListenAndServe()
		if err != http.ErrServerClosed {
			t.Errorf("unexpected error starting HTTP server: %v", err)
		}
	}()

	t.Run("Healthy", func(t *testing.T) {

		wants := &health.HTTPReport{
			Status: health.Healthy,
		}

		h := New()
		report := h.HTTP(48052)

		if report.Query == 0 {
			t.Errorf("expected query to take time to complete, cannot be zero")
			return
		}
		report.Query = 0

		if !reflect.DeepEqual(wants, report) {
			t.Errorf("output mismatch error: wanted %v ; got %v", wants, report)
		}
	})

	err := s200.Shutdown(context.Background())
	if err != nil {
		t.Errorf("unexpected error stopping HTTP server: %v", err)
	}

	s500 := httpServer(mock500)
	go func() {
		err := s500.ListenAndServe()
		if err != http.ErrServerClosed {
			t.Errorf("unexpected error starting HTTP server: %v", err)
		}

	}()

	t.Run("Healthy", func(t *testing.T) {

		wants := &health.HTTPReport{
			Status: health.Unhealthy,
		}

		h := New()
		report := h.HTTP(48052)

		if report.Query == 0 {
			t.Errorf("expected query to take time to complete, cannot be zero")
			return
		}
		report.Query = 0

		if !reflect.DeepEqual(wants, report) {
			t.Errorf("output mismatch error: wanted %v ; got %v", wants, report)
		}
	})

	err = s500.Shutdown(context.Background())
	if err != nil {
		t.Errorf("unexpected error stopping HTTP server: %v", err)
	}

	t.Run("Stopped", func(t *testing.T) {
		wants := &health.HTTPReport{
			Status: health.Stopped,
		}

		h := New()
		report := h.HTTP(48052)

		if report.Query == 0 {
			t.Errorf("expected query to take time to complete, cannot be zero")
			return
		}
		report.Query = 0

		if !reflect.DeepEqual(wants, report) {
			t.Errorf("output mismatch error: wanted %v ; got %v", wants, report)
		}
	})
}

func TestMerge(t *testing.T) {
	t.Run("Healthy", func(t *testing.T) {
		wants := health.Healthy

		h := New()
		report := h.Merge(
			&health.StoreReport{
				Status: health.Healthy,
			},
			&health.DNSReport{
				Status: health.Healthy,
			},
			&health.HTTPReport{
				Status: health.Healthy,
			},
		)

		if report.Status != wants {
			t.Errorf("output mismatch error: wanted %v ; got %v", wants, report.Status)
		}
	})
	t.Run("Running", func(t *testing.T) {
		wants := health.Running

		h := New()
		report := h.Merge(
			&health.StoreReport{
				Status: health.Healthy,
			},
			&health.DNSReport{
				Status: health.Healthy,
			},
			&health.HTTPReport{
				Status: health.Running,
			},
		)

		if report.Status != wants {
			t.Errorf("output mismatch error: wanted %v ; got %v", wants, report.Status)
		}
	})
	t.Run("Unhealthy", func(t *testing.T) {
		wants := health.Unhealthy

		h := New()
		report := h.Merge(
			&health.StoreReport{
				Status: health.Healthy,
			},
			&health.DNSReport{
				Status: health.Healthy,
			},
			&health.HTTPReport{
				Status: health.Unhealthy,
			},
		)

		if report.Status != wants {
			t.Errorf("output mismatch error: wanted %v ; got %v", wants, report.Status)
		}
	})
}
