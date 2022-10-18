package config

import "net"

type DNSConfig struct {
	Type        string `json:"type,omitempty" yaml:"type,omitempty"`
	FallbackDNS string `json:"fallback,omitempty" yaml:"fallback,omitempty"`
	Address     string `json:"address,omitempty" yaml:"address,omitempty"`
	Prefix      string `json:"prefix,omitempty" yaml:"prefix,omitempty"`
	Proto       string `json:"proto,omitempty" yaml:"proto,omitempty"`
}

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

func DNSFallback(f string) ConfigOption {
	return &dnsFallback{
		f: f,
	}
}

func DNSAddress(a string) ConfigOption {
	addr := net.ParseIP(a)
	if addr.IsUnspecified() {
		return nil
	}
	return &dnsAddress{
		a: a,
	}
}

func DNSPrefix(p string) ConfigOption {
	if len(p) > 1 {
		p = string(p[0])
	}
	return &dnsPrefix{
		p: p,
	}
}

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

func (l *dnsType) Apply(c *Config) {
	c.DNS.Type = l.t

}

func (l *dnsFallback) Apply(c *Config) {

	c.DNS.FallbackDNS = l.f
}

func (l *dnsAddress) Apply(c *Config) {
	c.DNS.Address = l.a
}

func (l *dnsPrefix) Apply(c *Config) {
	c.DNS.Prefix = l.p
}

func (l *dnsProto) Apply(c *Config) {
	c.DNS.Proto = l.p
}
