package config

import "os"

// ConfigType creates a ConfigOption setting the Config's type to string `t`
//
// It defaults to `yaml`
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

// ConfigPath creates a ConfigOption setting the Config's path to string `t`
//
// It tries to call os.Stat() on string `p` to evaluate if the file exists.
// If it doesn't, it attempts to create it. If that fails too the returned ConfigOption
// is `nil`; otherwise it will return a ConfigOption to update the config path.
func ConfigPath(p string) ConfigOption {
	_, err := os.Stat(p)
	if err != nil {
		f, err := os.Create(p)
		if err != nil {
			return nil
		}
		err = f.Sync()
		if err != nil {
			return nil
		}
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

// Apply implements the ConfigOption interface
func (l *configType) Apply(c *Config) {
	c.Type = l.t
}

// Apply implements the ConfigOption interface
func (l *configPath) Apply(c *Config) {
	c.Path = l.p
}
