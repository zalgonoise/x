package config

// Config structures the setup of the DNS app, according to the caller's needs
//
// This information can also be stored and loaded in a file for quicker access in
// future executions
type Config struct {
	DNS       *DNSConfig       `json:"dns,omitempty" yaml:"dns,omitempty"`
	Store     *StoreConfig     `json:"store,omitempty" yaml:"store,omitempty"`
	HTTP      *HTTPConfig      `json:"http,omitempty" yaml:"http,omitempty"`
	Logger    *LoggerConfig    `json:"logger,omitempty" yaml:"logger,omitempty"`
	Autostart *AutostartConfig `json:"autostart,omitempty" yaml:"autostart,omitempty"`
	Health    *HealthConfig    `json:"health,omitempty" yaml:"health,omitempty"`
	Type      string           `json:"type,omitempty" yaml:"type,omitempty"`
	Path      string           `json:"path,omitempty" yaml:"path,omitempty"`
}

// ConfigOption describes setter types for a Config
//
// As new options / elements are added to the Config, new data structures can
// implement the ConfigOption interface to allow setting these options in the Config
type ConfigOption interface {
	Apply(*Config)
}

// New initializes a new config with default settings, and then iterates through
// all input ConfigOption `opts` applying them to the Config, which is returned
// to the caller
func New(opts ...ConfigOption) *Config {
	conf := Default()

	for _, opt := range opts {
		if opt != nil {
			opt.Apply(conf)
		}
	}

	return conf
}

// Apply implements the ConfigOption interface
//
// It allows applying new options on top of an already existing config
func (c *Config) Apply(opts ...ConfigOption) *Config {
	for _, opt := range opts {
		if opt != nil {
			opt.Apply(c)
		}
	}
	return c
}

// Default returns a pointer to a Config, initialized with sane defaults
// and ready for automatic start-up
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
		Health: &HealthConfig{
			Type: "simplehealth",
		},
		Autostart: &AutostartConfig{
			DNS: true,
		},
	}
}

// Merge combines Configs `main` with `input`, returning a merged version
// of the two
//
// All set elements in `input` will be applied to `main`, and the unset elements
// will be ignored (keeps `main`'s data)
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
