package config

type Config struct {
	HTTPPort       int    `json:"http_port,omitempty" yaml:"http_port,omitempty"`
	BoltDBPath     string `json:"bolt_path,omitempty" yaml:"bolt_path,omitempty"`
	SQLiteDBPath   string `json:"sqlite_path,omitempty" yaml:"sqlite_path,omitempty"`
	SigningKeyPath string `json:"jwt_key_path,omitempty" yaml:"jwt_key_path,omitempty"`
	LogFilePath    string `json:"logfile_path,omitempty" yaml:"logfile_path,omitempty"`
	TraceFilePath  string `json:"tracefile_path,omitempty" yaml:"tracefile_path,omitempty"`
}

var Default = Config{
	HTTPPort:       8080,
	BoltDBPath:     "/secr/keys.db",
	SQLiteDBPath:   "/secr/sqlite.db",
	SigningKeyPath: "/secr/server/key",
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
func (c *Config) Apply(opts ...ConfigOption) *Config {
	for _, opt := range opts {
		if opt != nil {
			opt.Apply(c)
		}
	}
	return c
}

// Merge combines Configs `main` with `input`, returning a merged version
// of the two
//
// All set elements in `input` will be applied to `main`, and the unset elements
// will be ignored (keeps `main`'s data)
func Merge(main, input *Config) *Config {
	if input.HTTPPort != 0 {
		main.HTTPPort = input.HTTPPort
	}
	if input.BoltDBPath != "" {
		main.BoltDBPath = input.BoltDBPath
	}
	if input.SQLiteDBPath != "" {
		main.SQLiteDBPath = input.SQLiteDBPath
	}
	if input.SigningKeyPath != "" {
		main.SigningKeyPath = input.SigningKeyPath
	}
	if input.LogFilePath != "" {
		main.LogFilePath = input.LogFilePath
	}
	return main
}
