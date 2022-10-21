package core

import (
	"strings"

	"github.com/zalgonoise/x/dns/dns"
)

const (
	fallbackOneDot = "1.1.1.1:53"
	fallbackGoogle = "8.8.8.8:53"
	portSep        = ":"
	portDNS        = ":53"
)

var (
	defaultFallback = []string{
		fallbackOneDot,
		fallbackGoogle,
	}
)

// DNSCore adds a basic Answer / Fallback interaction for miekg's DNS
// implementation (used on the transport layer)
//
// It holds a list of strings which will be the fallback domain-name servers
// to contact if a domain's IP is requested but there no records for it
type DNSCore struct {
	fallbackDNS []string
}

// New returns a new DNSCore as a dns.Repository
func New(fallbackDNS ...string) dns.Repository {
	var fbDNS []string

	for _, fb := range fallbackDNS {
		fb := fb
		if fb == "" {
			continue
		}
		if !strings.Contains(fb, portSep) {
			fb += portDNS
		}
		fbDNS = append(fbDNS, fb)
	}

	if len(fbDNS) == 0 {
		// add defaults
		fbDNS = defaultFallback
	}

	return &DNSCore{
		fallbackDNS: fbDNS,
	}
}
