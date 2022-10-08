package dns

import (
	dnsr "github.com/miekg/dns"
)

type DNSRepository interface {
	ParseQuery(m *dnsr.Msg)
	HandleRequest(w dnsr.ResponseWriter, r *dnsr.Msg)
	Start() error
	Stop() error
	Reload() error
}
