package dns

type DNS struct {
	Addr   string
	Prefix string
}

type DNSBuilder struct {
	addr   string
	prefix string
}

func New() *DNSBuilder {
	return &DNSBuilder{}
}

func (b *DNSBuilder) Addr(s string) *DNSBuilder {
	b.addr = s
	return b
}

func (b *DNSBuilder) Prefix(s string) *DNSBuilder {
	b.prefix = s
	return b
}

func (b *DNSBuilder) Build() *DNS {
	if b.addr == "" {
		b.addr = ":53"
	}
	if b.prefix == "" {
		b.prefix = "."
	}
	return &DNS{
		Addr:   b.addr,
		Prefix: b.prefix,
	}
}
