package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type envConfig struct {
	// Mode sets the operation mode for the processor
	Mode string `envconfig:"X_AUDIO_MODE"`
	// URL points to an HTTP audio stream source
	URL string `envconfig:"X_AUDIO_URL"`
	// Duration delimits a stream's runtime duration
	Duration time.Duration `envconfig:"X_AUDIO_TIMEOUT"`
	// Output sets the type of Output for the processor
	Output string `envconfig:"X_AUDIO_OUTPUT_TYPE"`
	// OutputPath describes the path (or URL) for the set Output if applicable
	OutputPath string `envconfig:"X_AUDIO_OUTPUT_PATH"`
	// ExitCode forces a custom exit code on the processor when done or errored
	ExitCode int `envconfig:"X_AUDIO_EXIT_CODE"`
}

func (c *envConfig) Config() *Config {
	return &Config{
		Mode:       OpMode(c.Mode),
		URL:        c.URL,
		Duration:   c.Duration,
		Output:     Output(c.Output),
		OutputPath: c.OutputPath,
		ExitCode:   c.ExitCode,
	}
}

// FromEnv loads a Config from the set environment variables in the system
func FromEnv() (*Config, error) {
	var conf = new(envConfig)

	if err := envconfig.Process("", conf); err != nil {
		return nil, err
	}

	return conf.Config(), nil
}
