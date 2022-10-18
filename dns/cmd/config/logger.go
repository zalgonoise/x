package config

import "os"

type LoggerConfig struct {
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
}

func LoggerPath(p string) ConfigOption {
	_, err := os.Stat(p)
	if err != nil {
		return nil
	}
	return &loggerPath{
		p: p,
	}
}
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

func (l *loggerPath) Apply(c *Config) {
	c.Logger.Path = l.p
}

func (l *loggerType) Apply(c *Config) {
	c.Logger.Type = l.t
}
