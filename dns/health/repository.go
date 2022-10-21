package health

import (
	"time"

	"github.com/zalgonoise/x/dns/store"
)

type Repository interface {
	Store(int, time.Duration) *StoreReport
	DNS(address string, fallback string, records *store.Record) *DNSReport
	HTTP(address int) *HTTPReport
	Merge(*StoreReport, *DNSReport, *HTTPReport) *Report
}
