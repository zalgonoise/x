package config

type AutostartConfig struct {
	DNS bool `json:"dns,omitempty" yaml:"dns,omitempty"`
}

// AutostartDNS creates a ConfigOption which will set the UDP server (DNS) to
// start listening when the app is executed (alongside the HTTP server), or to
// wait for a HTTP request against /dns/start
func AutostartDNS(s bool) ConfigOption {
	return &autostartDNS{
		s: s,
	}
}

type autostartDNS struct {
	s bool
}

// Apply implements the ConfigOption interface
func (a *autostartDNS) Apply(c *Config) {
	c.Autostart.DNS = a.s
}
