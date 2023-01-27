package config

// Config describes the app configuration
type Config struct {
	HTTPPort       int    `json:"http_port,omitempty" yaml:"http_port,omitempty"`
	BoltDBPath     string `json:"bolt_path,omitempty" yaml:"bolt_path,omitempty"`
	SQLiteDBPath   string `json:"sqlite_path,omitempty" yaml:"sqlite_path,omitempty"`
	SigningKeyPath string `json:"jwt_key_path,omitempty" yaml:"jwt_key_path,omitempty"`
	LogFilePath    string `json:"logfile_path,omitempty" yaml:"logfile_path,omitempty"`
	TraceFilePath  string `json:"tracefile_path,omitempty" yaml:"tracefile_path,omitempty"`
}

// Default is a default configuration that the app will kick-off with, if not configured
var Default = Config{
	HTTPPort:       8080,
	BoltDBPath:     "/secr/keys.db",
	SQLiteDBPath:   "/secr/sqlite.db",
	SigningKeyPath: "/secr/server/key",
}

// Option describes setter types for a Config
//
// As new options / elements are added to the Config, new data structures can
// implement the Option interface to allow setting these options in the Config
type Option interface {
	// Apply sets the configuration on the input Config `c`
	Apply(c *Config)
}

// New initializes a new config with default settings, and then iterates through
// all input Option `opts` applying them to the Config, which is returned
// to the caller
func New(opts ...Option) *Config {
	conf := &Default

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
func (c *Config) Apply(opts ...Option) *Config {
	for _, opt := range opts {
		if opt != nil {
			opt.Apply(c)
		}
	}
	return c
}

// Merge combines Configs `c` with `input`, returning a merged version
// of the two
//
// All set elements in `input` will be applied to `c`, and the unset elements
// will be ignored (keeps `c`'s data)
func (c *Config) Merge(input *Config) *Config {
	if input.HTTPPort != 0 {
		c.HTTPPort = input.HTTPPort
	}
	if input.BoltDBPath != "" {
		c.BoltDBPath = input.BoltDBPath
	}
	if input.SQLiteDBPath != "" {
		c.SQLiteDBPath = input.SQLiteDBPath
	}
	if input.SigningKeyPath != "" {
		c.SigningKeyPath = input.SigningKeyPath
	}
	if input.LogFilePath != "" {
		c.LogFilePath = input.LogFilePath
	}
	return c
}
