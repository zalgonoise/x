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
	defaultDNSTarget     = "google.com"
	defaultDNSRecordType = "A"
)

type shealth struct{}

func New() health.Repository {
	return &shealth{}
}

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
	out.Duration = t

	return out
}

func (h *shealth) DNS(address string, fallback string, record *store.Record) *health.DNSReport {
	var (
		isFailing bool
		out       = &health.DNSReport{}
	)

	// internal query
	if record != nil {
		message := new(dnsr.Msg)
		message.SetQuestion(dnsr.Fqdn(record.Name), store.RecordTypeInts[record.Type])
		client := &dnsr.Client{
			DialTimeout:  2 * time.Second,
			ReadTimeout:  2 * time.Second,
			WriteTimeout: 2 * time.Second,
			Net:          "udp",
		}

		before := time.Now()
		_, _, err := client.Exchange(message, address)
		out.LocalQuery = time.Since(before)

		if err != nil {
			isFailing = true
		}
	}

	if fallback != "" {
		message := new(dnsr.Msg)
		message.SetQuestion(defaultDNSTarget, store.RecordTypeInts[defaultDNSRecordType])
		client := &dnsr.Client{
			DialTimeout:  2 * time.Second,
			ReadTimeout:  2 * time.Second,
			WriteTimeout: 2 * time.Second,
			Net:          "udp",
		}

		before := time.Now()
		_, _, err := client.Exchange(message, fallback)
		out.ExternalQuery = time.Since(before)

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

func (h *shealth) HTTP(port int) *health.HTTPReport {
	var out = &health.HTTPReport{}

	address := fmt.Sprintf("http://localhost:%v/reports", port)

	before := time.Now()
	res, err := http.Get(address)
	out.Query = time.Since(before)
	if err != nil {
		out.Status = health.Stopped
	} else if res.StatusCode > 399 {
		out.Status = health.Unhealthy
	} else if res.StatusCode == 200 {
		out.Status = health.Healthy
	}

	return out
}

func (h *shealth) Merge(
	storeH *health.StoreReport,
	dnsH *health.DNSReport,
	httpH *health.HTTPReport,
) *health.Report {
	out := &health.Report{}

	if storeH.Status > 1 && dnsH.Status > 1 && httpH.Status > 1 {
		out.Status = health.Running
		if storeH.Status == health.Healthy &&
			dnsH.Status == health.Healthy &&
			httpH.Status == health.Healthy {
			out.Status = health.Healthy
		}
	}

	out.StoreReport = storeH
	out.DNSReport = dnsH
	out.HTTPReport = httpH

	return out
}
