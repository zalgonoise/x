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
			Type:    "miekgdns",
			Address: ":53",
			Prefix:  ".",
			Proto:   "udp",
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
