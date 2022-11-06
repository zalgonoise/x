package dns

import (
	"github.com/miekg/dns"
	"github.com/zalgonoise/x/dns/store"
)

// Repository defines the set of operations that a DNS answerer should expose
//
// This will consist in building answers for DNS messages based on store.Records,
// or by fetching the answers from a fallback / secondary DNS
type Repository interface {
	// Answer will write the IP address present in the store.Record in the dns.Msg
	// slice of Answers
	Answer(*store.Record, *dns.Msg)
	// Fallback is called when the DNS store does not hold a record for the requested
	// domain, so the DNS service spawns a DNS client that will query the fallback server
	// and write that answer to the dns.Msg
	Fallback(*store.Record, *dns.Msg)
}
