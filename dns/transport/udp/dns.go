package udp

const (
	addr   string = ":53"
	proto  string = "udp"
	prefix string = "."
)

type DNS struct {
	Addr   string
	Prefix string
	Proto  string
}

type DNSBuilder struct {
	addr   string
	prefix string
	proto  string
}

func NewDNS() *DNSBuilder {
	return &DNSBuilder{}
}

func (b *DNSBuilder) Addr(s string) *DNSBuilder {
	b.addr = s
	return b
}

func (b *DNSBuilder) Prefix(s string) *DNSBuilder {
	// input is string, but we're looking for a rune
	if len(s) > 1 {
		s = string(s[0])
	}
	b.prefix = s
	return b
}

func (b *DNSBuilder) Proto(s string) *DNSBuilder {
	b.proto = s
	return b
}

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
