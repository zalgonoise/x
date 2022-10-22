package health

import (
	"time"

	"github.com/zalgonoise/x/dns/store"
)

// Repository defines the set of operations that a health checker should have
//
// This will consist in basic operations to retrieve the status (and other metadata)
// from the running modules and services, while also being able to merge all of these
// individual reports into one
type Repository interface {
	// Store will take in the number of records in the store and the time.Duration for a
	// store.List operation, and return a StoreReport based off of this information
	Store(int, time.Duration) *StoreReport
	// DNS will take in the address of the UDP server, the fallback DNS address (if set),
	// and a store.Record, which are used to answer internal and external DNS queries as part
	// of a health check; returning a DNSReport based off of this information
	DNS(address string, fallback string, records *store.Record) *DNSReport
	// HTTP will take the HTTP server's port so it can perform a HTTP request against one
	// of its endpoints, and returning a HTTPReport based off of this information
	HTTP(port int) *HTTPReport
	// Merge will unite a StoreReport, DNSReport and HTTPReport, returning a Report which
	// encapsulates these as well as an overall status for the service
	Merge(*StoreReport, *DNSReport, *HTTPReport) *Report
}
