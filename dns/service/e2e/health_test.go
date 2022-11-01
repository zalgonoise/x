package e2e

import (
	"context"
	"reflect"
	"testing"

	"github.com/zalgonoise/x/dns/health"
)

func TestHealth(t *testing.T) {
	s := initializeService()

	err := s.AddRecords(context.Background(), record1, record2)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	t.Run("StoreHealth", func(t *testing.T) {
		report := s.StoreHealth()
		wants := &health.StoreReport{
			Len:    2,
			Status: health.Healthy,
		}

		if report == nil {
			t.Errorf("expected output not to be nil")
			return
		}
		if report.Duration == 0 {
			t.Errorf("op duration value cannot be zero")
			return
		}
		report.Duration = 0

		if !reflect.DeepEqual(wants, report) {
			t.Errorf("output mismatch error: wanted %v ; got %v", wants, report)
		}
	})

	t.Run("DNSHealth", func(t *testing.T) {
		report := s.DNSHealth()
		wants := &health.DNSReport{
			Enabled: false,
			Status:  health.Unhealthy,
		}

		if report == nil {
			t.Errorf("expected output not to be nil")
			return
		}
		if report.LocalQuery == 0 {
			t.Errorf("local op duration value cannot be zero")
			return
		}
		if report.ExternalQuery == 0 {
			t.Errorf("external op duration value cannot be zero")
			return
		}
		report.LocalQuery = 0
		report.ExternalQuery = 0

		if !reflect.DeepEqual(wants, report) {
			t.Errorf("output mismatch error: wanted %v ; got %v", wants, report)
		}
	})

	t.Run("HTTPHealth", func(t *testing.T) {
		report := s.HTTPHealth()
		wants := &health.HTTPReport{
			Status: health.Stopped,
		}

		if report == nil {
			t.Errorf("expected output not to be nil")
			return
		}
		if report.Query == 0 {
			t.Errorf("op duration value cannot be zero")
			return
		}
		report.Query = 0

		if !reflect.DeepEqual(wants, report) {
			t.Errorf("output mismatch error: wanted %v ; got %v", wants, report)
		}
	})

	t.Run("Health", func(t *testing.T) {
		report := s.Health()
		wantsStore := &health.StoreReport{
			Len:    2,
			Status: health.Healthy,
		}
		wantsDNS := &health.DNSReport{
			Enabled: false,
			Status:  health.Unhealthy,
		}
		wantsHTTP := &health.HTTPReport{
			Status: health.Stopped,
		}
		if report == nil {
			t.Errorf("expected output not to be nil")
			return
		}
		if report.StoreReport == nil || report.DNSReport == nil || report.HTTPReport == nil {
			t.Errorf("inner report cannot be nil")
			return
		}

		if report.StoreReport.Duration == 0 {
			t.Errorf("store op duration value cannot be nil")
			return
		}
		if report.DNSReport.LocalQuery == 0 || report.DNSReport.ExternalQuery == 0 {
			t.Errorf("dns op duration value cannot be nil")
			return
		}
		if report.HTTPReport.Query == 0 {
			t.Errorf("http op duration value cannot be nil")
			return
		}

		report.StoreReport.Duration = 0
		report.DNSReport.LocalQuery = 0
		report.DNSReport.ExternalQuery = 0
		report.HTTPReport.Query = 0

		if !reflect.DeepEqual(wantsStore, report.StoreReport) {
			t.Errorf("output mismatch error: wanted %v ; got %v", wantsStore, report.StoreReport)
		}
		if !reflect.DeepEqual(wantsDNS, report.DNSReport) {
			t.Errorf("output mismatch error: wanted %v ; got %v", wantsDNS, report.DNSReport)
		}
		if !reflect.DeepEqual(wantsHTTP, report.HTTPReport) {
			t.Errorf("output mismatch error: wanted %v ; got %v", wantsHTTP, report.HTTPReport)
		}
		if report.Status != health.Unhealthy {
			t.Errorf("output mimatch error: wanted %v ; got %v", health.Unhealthy, report.Status)
		}
	})
}
