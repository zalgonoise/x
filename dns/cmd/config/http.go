package config

type HTTPConfig struct {
	Port int `json:"port,omitempty" yaml:"port,omitempty"`
}

func HTTPPort(p int) ConfigOption {
	if p > 65535 {
		return nil
	}
	return &httpPort{
		p: p,
	}
}

type httpPort struct {
	p int
}

func (h *httpPort) Apply(c *Config) {
	c.HTTP.Port = h.p
}
