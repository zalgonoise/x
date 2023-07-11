package config

import (
	"flag"
)

func NewConfig() (*Config, error) {
	url := flag.String("url", "", "the URL for the WAV audio stream")
	mode := flag.String("mode", "monitor", "defines the operation mode")
	out := flag.String("to", "logger", "defines the output mode [logger, file, prometheus]")
	path := flag.String("o", "", "defines the path to the output (if a file, or a port / address for Prometheus)")
	exit := flag.Int("exit", 0, "sets a custom exit code for when the app exits")

	flag.Parse()

	config := &Config{
		Mode:       OpMode(*mode),
		URL:        *url,
		Output:     Output(*out),
		OutputPath: *path,
		ExitCode:   *exit,
	}

	return config, Validate(config)
}
