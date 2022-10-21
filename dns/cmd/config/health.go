package config

type HealthConfig struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
}

func HealthType(t string) ConfigOption {
	switch t {
	case "simplehealth", "simple":
		return &healthType{
			t: "simplehealth",
		}
	default:
		return &healthType{
			t: "simplehealth",
		}
	}
}

type healthType struct {
	t string
}

func (l *healthType) Apply(c *Config) {
	c.Health.Type = l.t
}
