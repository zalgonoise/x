package core

import (
	"fmt"

	dnsr "github.com/miekg/dns"
	"github.com/zalgonoise/x/dns/store"
	// "github.com/zalgonoise/zlog/log/event"
)

// DNSCore adds a basic Answer interaction for miekg's DNS, used by
// the service
type DNSCore struct {
	err error
}

func New() *DNSCore {
	return &DNSCore{}
}

func (d *DNSCore) Answer(r *store.Record, m *dnsr.Msg) {
	response, err := dnsr.NewRR(
		fmt.Sprintf("%s %s %s", r.Name, r.Type, r.Addr),
	)
	if err != nil {
		d.err = err
		return
	}
	m.Answer = append(m.Answer, response)
}
