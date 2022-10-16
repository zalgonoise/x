package core

import (
	"fmt"
	"strings"
	"time"

	dnsr "github.com/miekg/dns"
	"github.com/zalgonoise/x/dns/store"
)

const (
	fallbackOneDot = "1.1.1.1:53"
	fallbackGoogle = "8.8.8.8:53"
	portSep        = ":"
	portDNS        = ":53"
)

// DNSCore adds a basic Answer interaction for miekg's DNS, used by
// the service
type DNSCore struct {
	fallbackDNS []string
}

func New(fallbackDNS ...string) *DNSCore {
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
		fbDNS = append(fbDNS, fallbackOneDot, fallbackGoogle)
	}

	return &DNSCore{
		fallbackDNS: fallbackDNS,
	}
}

func (d *DNSCore) Answer(r *store.Record, m *dnsr.Msg) {
	response, err := dnsr.NewRR(
		fmt.Sprintf("%s %s %s", r.Name, r.Type, r.Addr),
	)
	if err != nil {
		return
	}
	m.Answer = append(m.Answer, response)
}

func (d *DNSCore) Fallback(r *store.Record, m *dnsr.Msg) {
	message := new(dnsr.Msg)
	message.SetQuestion(dnsr.Fqdn(r.Name), store.RecordTypeInts[r.Type])
	client := &dnsr.Client{
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		Net:          "udp",
	}

	for _, fallback := range d.fallbackDNS {
		in, _, err := client.Exchange(message, fallback)
		if err != nil || len(in.Answer) == 0 {
			continue
		}
		m.Answer = append(m.Answer, in.Answer...)
		break
	}
}
