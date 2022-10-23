package core

import (
	"fmt"
	"time"

	dns "github.com/miekg/dns"
	"github.com/zalgonoise/x/dns/store"
)

// Answer will take the IP address populated in the store.Response `r`, and
// append it as a DNS response to the dns.Msg `m`'s Answer slice
func (d *DNSCore) Answer(r *store.Record, m *dns.Msg) {
	response, err := dns.NewRR(
		fmt.Sprintf("%s %s %s", r.Name, r.Type, r.Addr),
	)
	if err != nil {
		return
	}
	m.Answer = append(m.Answer, response)
}

// Fallback will spawn a DNS client and issues a request to the fallback servers
// with the same query for which there isn't a record in the store.
//
// If there is an answer, it is written to the dns.Msg `r`'s Answer slice; otherwise
// the request is discarted until it times out
func (d *DNSCore) Fallback(r *store.Record, m *dns.Msg) {
	message := new(dns.Msg)
	message.SetQuestion(dns.Fqdn(r.Name), store.RecordTypeInts[r.Type])
	client := &dns.Client{
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
