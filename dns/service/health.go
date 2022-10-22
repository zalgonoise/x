package service

import (
	"context"
	"strings"
	"time"

	"github.com/zalgonoise/x/dns/health"
	"github.com/zalgonoise/x/dns/store"
)

// StoreHealth uses the health.Repository to generate a health.StoreReport
func (s *service) StoreHealth() *health.StoreReport {
	before := time.Now()
	r, err := s.store.List(context.Background())
	after := time.Since(before)
	if err != nil {
		return s.health.Store(0, 0)
	}
	return s.health.Store(len(r), after)
}

// DNSHealth uses the health.Repository to generate a health.DNSReport
func (s *service) DNSHealth() *health.DNSReport {
	var addr string

	r, err := s.store.List(context.Background())
	if err != nil || len(r) == 0 {
		r = []*store.Record{nil}
	}

	if s.conf.DNS.FallbackDNS != "" {
		addr = strings.Split(s.conf.DNS.FallbackDNS, ",")[0]
	}

	return s.health.DNS(
		s.conf.DNS.Address,
		addr,
		r[0],
	)

}

// HTTPHealth uses the health.Repository to generate a health.HTTPReport
func (s *service) HTTPHealth() *health.HTTPReport {
	return s.health.HTTP(s.conf.HTTP.Port)
}

// Health uses the health.Repository to generate a health.Report
func (s *service) Health() *health.Report {
	return s.health.Merge(
		s.StoreHealth(),
		s.DNSHealth(),
		s.HTTPHealth(),
	)
}
