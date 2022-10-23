package config

import "os"

type LoggerConfig struct {
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
}

// LoggerPath creates a ConfigOption setting the Config's logfile path to string `p`
//
// It tries to call os.Stat() on string `p` to evaluate if the file exists.
// If it doesn't, it attempts to create it. If that fails too the returned ConfigOption
// is `nil`; otherwise it will return a ConfigOption to update the store path.
func LoggerPath(p string) ConfigOption {
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
	return &loggerPath{
		p: p,
	}
}

// LoggerType creates a ConfigOption setting the Config's logger type to string `t`
//
// It defaults to `text`
func LoggerType(t string) ConfigOption {
	switch t {
	case "none":
		return &loggerType{
			t: "",
		}
	case "json", "text":
		return &loggerType{
			t: t,
		}
	default:
		return &loggerType{
			t: "text",
		}
	}
}

type loggerPath struct {
	p string
}
type loggerType struct {
	t string
}

// Apply implements the ConfigOption interface
func (l *loggerPath) Apply(c *Config) {
	c.Logger.Path = l.p
}

// Apply implements the ConfigOption interface
func (l *loggerType) Apply(c *Config) {
	c.Logger.Type = l.t
}
