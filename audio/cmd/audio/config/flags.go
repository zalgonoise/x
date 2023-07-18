package config

import (
	"flag"
	"time"
)

// NewConfig creates a new Config by reading the input flags to the application startup
//
// It returns a new Config and an error, which is a call to the Validate(Config) function
func NewConfig() (*Config, error) {
	url := flag.String("url", "", "the URL for the WAV audio stream")
	mode := flag.String("mode", "monitor", "defines the operation mode")
	out := flag.String("to", "logger", "defines the output mode [logger, file, prometheus]")
	path := flag.String("o", "", "defines the path to the output (if a file, or a port / address for Prometheus)")
	exit := flag.Int("exit", 0, "sets a custom exit code for when the app exits")
	timeout := flag.String("dur", "30s", "sets the duration of the recording or analysis")

	flag.Parse()

	var dur time.Duration
	var err error

	dur, err = time.ParseDuration(*timeout)
	if err != nil || dur < 0 {
		dur = 0
	}

	config := &Config{
		Mode:       OpMode(*mode),
		URL:        *url,
		Duration:   dur,
		Output:     Output(*out),
		OutputPath: *path,
		ExitCode:   *exit,
	}

	return config, Validate(config)
}
