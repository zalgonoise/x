package simplehealth

import (
	"fmt"
	"net/http"
	"time"

	dnsr "github.com/miekg/dns"
	"github.com/zalgonoise/x/dns/health"
	"github.com/zalgonoise/x/dns/store"
)

const (
	defaultDNSTarget     = "google.com."
	defaultDNSRecordType = "A"
)

type shealth struct{}

// New will return a new shealth as a health.Repository
func New() health.Repository {
	return &shealth{}
}

// Store will take in the number of records in the store and the time.Duration for a
// store.List operation, and return a StoreReport based off of this information
func (h *shealth) Store(length int, t time.Duration) *health.StoreReport {
	var out = &health.StoreReport{}

	switch length {
	case 0:
		if t == 0 {
			out.Status = health.Stopped
			return out
		}
		out.Status = health.Running
	default:
		out.Status = health.Healthy
	}

	out.Len = length
	out.Duration = float64(t.Nanoseconds()) / 1000000

	return out
}

// DNS will take in the address of the UDP server, the fallback DNS address (if set),
// and a store.Record, which are used to answer internal and external DNS queries as part
// of a health check; returning a DNSReport based off of this information
func (h *shealth) DNS(address string, fallback string, record *store.Record) *health.DNSReport {
	var (
		isFailing bool
		out       = &health.DNSReport{}
		client    = &dnsr.Client{
			DialTimeout:  2 * time.Second,
			ReadTimeout:  2 * time.Second,
			WriteTimeout: 2 * time.Second,
			Net:          "udp",
		}
	)

	// internal query
	if record != nil {
		message := new(dnsr.Msg)
		message.SetQuestion(dnsr.Fqdn(record.Name), store.RecordTypeInts[record.Type])

		before := time.Now()
		_, _, err := client.Exchange(message, address)
		out.LocalQuery = float64(time.Since(before).Nanoseconds()) / 1000000

		if err != nil {
			isFailing = true
		}
	}

	if fallback != "" {
		message := new(dnsr.Msg)
		message.SetQuestion(defaultDNSTarget, store.RecordTypeInts[defaultDNSRecordType])

		before := time.Now()
		_, _, err := client.Exchange(message, address)
		out.ExternalQuery = float64(time.Since(before).Nanoseconds()) / 1000000

		if err != nil {
			isFailing = true
		}
	}

	switch isFailing {
	case true:
		out.Status = health.Unhealthy
		return out
	default:
		if out.LocalQuery == 0 && out.ExternalQuery == 0 {
			out.Enabled = false
			out.Status = health.Stopped
			return out
		}
		out.Enabled = true
		out.Status = health.Healthy

	}

	return out
}

// HTTP will take the HTTP server's port so it can perform a HTTP request against one
// of its endpoints, and returning a HTTPReport based off of this information
func (h *shealth) HTTP(port int) *health.HTTPReport {
	var out = &health.HTTPReport{}

	address := fmt.Sprintf("http://localhost:%v/records", port)

	before := time.Now()
	res, err := http.Get(address)
	out.Query = float64(time.Since(before).Nanoseconds()) / 1000000
	if err != nil {
		out.Status = health.Stopped
	} else if res.StatusCode > 399 {
		out.Status = health.Unhealthy
	} else if res.StatusCode == 200 {
		out.Status = health.Healthy
	}

	return out
}

// Merge will unite a StoreReport, DNSReport and HTTPReport, returning a Report which
// encapsulates these as well as an overall status for the service
func (h *shealth) Merge(
	storeH *health.StoreReport,
	dnsH *health.DNSReport,
	httpH *health.HTTPReport,
) *health.Report {
	out := &health.Report{}

	if (storeH.Status == health.Running || storeH.Status == health.Healthy) &&
		(dnsH.Status == health.Running || dnsH.Status == health.Healthy) &&
		(httpH.Status == health.Running || httpH.Status == health.Healthy) {
		out.Status = health.Running
		if storeH.Status == health.Healthy &&
			dnsH.Status == health.Healthy &&
			httpH.Status == health.Healthy {
			out.Status = health.Healthy
		}
	} else {
		out.Status = health.Unhealthy
	}

	out.StoreReport = storeH
	out.DNSReport = dnsH
	out.HTTPReport = httpH

	return out
}
