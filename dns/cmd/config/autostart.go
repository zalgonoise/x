package config

type AutostartConfig struct {
	DNS bool `json:"dns,omitempty" yaml:"dns,omitempty"`
}

func AutostartDNS(s bool) ConfigOption {
	return &autostartDNS{
		s: s,
	}
}

type autostartDNS struct {
	s bool
}

func (a *autostartDNS) Apply(c *Config) {
	c.Autostart.DNS = a.s
}
