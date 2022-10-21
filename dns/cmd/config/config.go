package config

import "os"

type Config struct {
	DNS       *DNSConfig       `json:"dns,omitempty" yaml:"dns,omitempty"`
	Store     *StoreConfig     `json:"store,omitempty" yaml:"store,omitempty"`
	HTTP      *HTTPConfig      `json:"http,omitempty" yaml:"http,omitempty"`
	Logger    *LoggerConfig    `json:"logger,omitempty" yaml:"logger,omitempty"`
	Autostart *AutostartConfig `json:"autostart,omitempty" yaml:"autostart,omitempty"`
	Type      string           `json:"type,omitempty" yaml:"type,omitempty"`
	Path      string           `json:"path,omitempty" yaml:"path,omitempty"`
}

func ConfigType(t string) ConfigOption {
	switch t {
	case "json", "yaml":
		return &configType{
			t: t,
		}
	default:
		return &configType{
			t: "yaml",
		}
	}
}

func ConfigPath(p string) ConfigOption {
	_, err := os.Stat(p)
	if err != nil {
		return nil
	}
	return &configPath{
		p: p,
	}
}

type configType struct {
	t string
}

type configPath struct {
	p string
}

func (l *configType) Apply(c *Config) {
	c.Type = l.t
}

func (l *configPath) Apply(c *Config) {
	c.Path = l.p
}

type ConfigOption interface {
	Apply(*Config)
}

func (c *Config) Apply(opts ...ConfigOption) *Config {
	for _, opt := range opts {
		if opt != nil {
			opt.Apply(c)
		}
	}

	return c
}

func Default() *Config {
	return &Config{
		DNS: &DNSConfig{
			Type:        "miekgdns",
			Address:     ":53",
			Prefix:      ".",
			Proto:       "udp",
			FallbackDNS: "1.1.1.1",
		},
		Store: &StoreConfig{
			Type: "memmap",
		},
		HTTP: &HTTPConfig{
			Port: 8080,
		},
		Logger: &LoggerConfig{
			Type: "text",
		},
		Autostart: &AutostartConfig{
			DNS: true,
		},
	}
}

func New(opts ...ConfigOption) *Config {
	conf := Default()

	for _, opt := range opts {
		if opt != nil {
			opt.Apply(conf)
		}
	}

	return conf
}

func Merge(main, input *Config) *Config {
	// DNS
	if input.DNS.Type != "" {
		main.DNS.Type = input.DNS.Type
	}
	if input.DNS.Address != "" {
		main.DNS.Address = input.DNS.Address
	}
	if input.DNS.Prefix != "" {
		main.DNS.Prefix = input.DNS.Prefix
	}
	if input.DNS.Proto != "" {
		main.DNS.Proto = input.DNS.Proto
	}
	if input.DNS.FallbackDNS != "" {
		main.DNS.FallbackDNS = input.DNS.FallbackDNS
	}

	// Store
	if input.Store.Type != "" {
		main.Store.Type = input.Store.Type
	}
	if input.Store.Path != "" {
		main.Store.Path = input.Store.Path
	}

	// HTTP
	if input.HTTP.Port != 0 {
		main.HTTP.Port = input.HTTP.Port
	}

	// Logger
	if input.Logger.Type != "" {
		main.Logger.Type = input.Logger.Type
	}
	if input.Logger.Path != "" {
		main.Logger.Path = input.Logger.Path
	}

	// Autostart
	if input.Autostart.DNS {
		main.Autostart.DNS = input.Autostart.DNS
	}

	if input.Path != "" {
		main.Path = input.Path
	}
	return main
}
