package config

type HTTPConfig struct {
	Port int `json:"port,omitempty" yaml:"port,omitempty"`
}

// HTTPPort creates a ConfigOption setting the Config's HTTP port to int `p`
//
// It returns nil if the input port `p` is 0 or over 65535, otherwise it will
// return a ConfigOption to update the HTTP port
func HTTPPort(p int) ConfigOption {
	if p > 65535 || p == 0 {
		return nil
	}
	return &httpPort{
		p: p,
	}
}

type httpPort struct {
	p int
}

// Apply implements the ConfigOption interface
func (h *httpPort) Apply(c *Config) {
	c.HTTP.Port = h.p
}
