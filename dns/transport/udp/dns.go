package udp

const (
	addr   string = ":53"
	proto  string = "udp"
	prefix string = "."
)

// DNS defines the structure of a DNS server, composed of its
// address, prefix character, and protocol
type DNS struct {
	Addr   string
	Prefix string
	Proto  string
}

// DNSBuilder is a builder type for DNS, allowing method chaining to
// set different properties, ended by a .Build() call
type DNSBuilder struct {
	addr   string
	prefix string
	proto  string
}

// NewDNS returns a new DNSBuilder
func NewDNS() *DNSBuilder {
	return &DNSBuilder{}
}

// Addr sets the DNS address as IP:Port (defaults to ":53")
func (b *DNSBuilder) Addr(s string) *DNSBuilder {
	b.addr = s
	return b
}

// Prefix sets the prefix character to append to added domains
// (defaults to ".")
//
// It is expected for the input string to be 1 character long,
// and this is enforced by checking if the length is greater than 1
// and taking the first rune of the string converted as a string
//
// When querying for "sub.domain.net", a DNS server expects to find
// a record like "sub.domain.net."
func (b *DNSBuilder) Prefix(s string) *DNSBuilder {
	// input is string, but we're looking for a rune
	if len(s) > 1 {
		s = string(s[0])
	}
	b.prefix = s
	return b
}

// Proto sets the protocol used for the DNS server (defaults to "udp")
func (b *DNSBuilder) Proto(s string) *DNSBuilder {
	b.proto = s
	return b
}

// Build will return a DNS based on the defined configuration, with
// defaults applied where unset
func (b *DNSBuilder) Build() *DNS {
	if b.addr == "" {
		b.addr = addr
	}
	if b.prefix == "" {
		b.prefix = prefix
	}
	if b.proto == "" {
		b.proto = proto
	}
	return &DNS{
		Addr:   b.addr,
		Prefix: b.prefix,
		Proto:  b.proto,
	}
}
