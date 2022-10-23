package config

import "net"

type DNSConfig struct {
	Type        string `json:"type,omitempty" yaml:"type,omitempty"`
	FallbackDNS string `json:"fallback,omitempty" yaml:"fallback,omitempty"`
	Address     string `json:"address,omitempty" yaml:"address,omitempty"`
	Prefix      string `json:"prefix,omitempty" yaml:"prefix,omitempty"`
	Proto       string `json:"proto,omitempty" yaml:"proto,omitempty"`
}

// DNSType creates a ConfigOption setting the Config's DNS type to string `t`
//
// It defaults to `miekgdns`
func DNSType(p string) ConfigOption {
	switch p {
	case "miekgdns", "miekg":
		return &dnsType{
			t: "miekgdns",
		}
	default:
		return &dnsType{
			t: "miekgdns",
		}
	}
}

// DNSFallback creates a ConfigOption setting the Config's fallback DNS address(es)
// to string `f`
//
// It the string `f` is empty, it returns `nil`
func DNSFallback(f string) ConfigOption {
	if f == "" {
		return nil
	}
	return &dnsFallback{
		f: f,
	}
}

// DNSAddress creates a ConfigOption setting the Config's DNS address to string `a`
//
// It the string `a` is an invalid IP address, it returns `nil`
func DNSAddress(a string) ConfigOption {
	addr := net.ParseIP(a)
	if addr.IsUnspecified() {
		return nil
	}
	return &dnsAddress{
		a: a,
	}
}

// DNSPrefix creates a ConfigOption setting the Config's DNS prefix to string `p`
//
// It the string `p` is longer than one character, the first rune is converted to a
// string and that one is used.
//
// DNS Prefix is a character inserted after a (simple, DNS store) domain which is
// required to perform the query, usually a dot (".")
//
// E.g.: if you store "dns.example.com", a query for it would ask for "dns.example.com."
func DNSPrefix(p string) ConfigOption {
	if len(p) > 1 {
		p = string(p[0])
	}
	return &dnsPrefix{
		p: p,
	}
}

// DNSProto creates a ConfigOption setting the Config's DNS proto to string `p`
//
// It defaults to `udp`
func DNSProto(p string) ConfigOption {
	switch p {
	case "udp":
		return &dnsProto{
			p: p,
		}
	default:
		return &dnsProto{
			p: "udp",
		}
	}
}

type dnsType struct {
	t string
}
type dnsFallback struct {
	f string
}
type dnsAddress struct {
	a string
}
type dnsPrefix struct {
	p string
}
type dnsProto struct {
	p string
}

// Apply implements the ConfigOption interface
func (l *dnsType) Apply(c *Config) {
	c.DNS.Type = l.t

}

// Apply implements the ConfigOption interface
func (l *dnsFallback) Apply(c *Config) {
	c.DNS.FallbackDNS = l.f
}

// Apply implements the ConfigOption interface
func (l *dnsAddress) Apply(c *Config) {
	c.DNS.Address = l.a
}

// Apply implements the ConfigOption interface
func (l *dnsPrefix) Apply(c *Config) {
	c.DNS.Prefix = l.p
}

// Apply implements the ConfigOption interface
func (l *dnsProto) Apply(c *Config) {
	c.DNS.Proto = l.p
}
