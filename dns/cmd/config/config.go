package config

import "os"

type Config struct {
	DNS       *DNSConfig       `json:"dns,omitempty" yaml:"dns,omitempty"`
	Store     *StoreConfig     `json:"store,omitempty" yaml:"store,omitempty"`
	HTTP      *HTTPConfig      `json:"http,omitempty" yaml:"http,omitempty"`
	Logger    *LoggerConfig    `json:"logger,omitempty" yaml:"logger,omitempty"`
	Autostart *AutostartConfig `json:"autostart,omitempty" yaml:"autostart,omitempty"`
	Path      string
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

type configPath struct {
	p string
}

func (l *configPath) Apply(c *Config) {
	c.Path = l.p
}

type ConfigOption interface {
	Apply(*Config)
}

func New(opts ...ConfigOption) *Config {
	conf := &Config{
		DNS:       &DNSConfig{},
		Store:     &StoreConfig{},
		HTTP:      &HTTPConfig{},
		Logger:    &LoggerConfig{},
		Autostart: &AutostartConfig{},
	}

	for _, opt := range opts {
		if opt != nil {
			opt.Apply(conf)
		}
	}

	return conf
}
