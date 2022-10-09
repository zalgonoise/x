package dns

import (
	dnsr "github.com/miekg/dns"
	"github.com/zalgonoise/x/dns/store"
)

// DNSRepository defines the behavior that a DNS Server should have
//
// This will consist in the basic UDP request handler actions as well as
// start / stop / reload functionalities.
//
// While the basic implementation is 100% based on `miekg/dns`, it is also
// possible to further extend the service with different implementations
type Repository interface {
	// ParseQuery will parse the incoming dns.Msg and append an answer
	// to it
	ParseQuery(m *dnsr.Msg)

	// HandleRequest is the dns.HandleFunc for a DNS server
	HandleRequest(w dnsr.ResponseWriter, r *dnsr.Msg)

	// Start will (re)launch the DNS Server
	Start() error

	// Stop will gracefully terminate the running DNS Server
	Stop() error

	// Reload will relaunch the running DNS server, taking into account
	// any Records updates in the Records Store
	Reload() error

	Store(store.Repository)
}
