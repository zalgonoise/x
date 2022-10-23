package config

type HealthConfig struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
}

// HealthType creates a ConfigOption setting the Config's health check type to string `t`
//
// It defaults to `simplehealth`
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

// Apply implements the ConfigOption interface
func (l *healthType) Apply(c *Config) {
	c.Health.Type = l.t
}
