package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type ServiceConfig struct {
	Threshold     int           `envconfig:"X_OBSERVABILITY_THRESH" default:"100"`
	MaxInputValue int           `envconfig:"X_MAX_INPUT_VALUE" default:"130"`
	Duration      time.Duration `envconfig:"X_OBSERVABILITY_DUR" default:"30s"`

	TracerURI   string `envconfig:"X_OBSERVABILITY_TRACER_URI" default:"localhost:4317"`
	MetricsPort int    `enconfig:"X_OBSERVABILITY_METRICS_PORT" default:"13090"`
}

func NewServiceConfig() (*ServiceConfig, error) {
	var c = &ServiceConfig{}

	if err := envconfig.Process("", c); err != nil {
		return nil, err
	}

	return c, nil
}
