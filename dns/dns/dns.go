package dns

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

func New() *DNSBuilder {
	return &DNSBuilder{}
}

func From(d *DNS) *DNSBuilder {
	return &DNSBuilder{
		addr:   d.Addr,
		prefix: d.Prefix,
		proto:  d.Proto,
	}
}

func (b *DNSBuilder) Addr(s string) *DNSBuilder {
	b.addr = s
	return b
}

func (b *DNSBuilder) Prefix(s string) *DNSBuilder {
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
