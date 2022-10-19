package service

import (
	"time"

	"github.com/zalgonoise/x/dns/health"
)

func (s *service) StoreHealth(entries int, t time.Duration) *health.Report {
	return nil
}
func (s *service) DNSHealth() *health.Report {
	return nil
}
