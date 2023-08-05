package config

import (
	"flag"
	"time"
)

var emptyFlagsConfig = Config{}

// FromFlags loads a Config from the input command-line flags, in the app's runtime
func FromFlags() *Config {
	mode := flag.String("mode", "", "defines the operation mode")
	url := flag.String("url", "", "the URL for the WAV audio stream")
	timeout := flag.String("dur", "", "sets the duration of the recording or analysis")
	out := flag.String("to", "", "defines the output mode [logger, file, prometheus]")
	path := flag.String("o", "", "defines the path to the output (if a file, or a port / address for Prometheus)")
	numBuckets := flag.Int("num-buckets", 0, "defines the number of buckets to separate frequencies in a spectrum analysis")
	exit := flag.Int("exit", 0, "sets a custom exit code for when the app exits")

	flag.Parse()

	var duration time.Duration
	if dur, err := time.ParseDuration(*timeout); err == nil && dur > 0 {
		duration = dur
	}

	config := &Config{
		Mode:               OpMode(*mode),
		URL:                *url,
		Duration:           duration,
		Output:             Output(*out),
		OutputPath:         *path,
		NumSpectrumBuckets: *numBuckets,
		ExitCode:           *exit,
	}

	if *config == emptyFlagsConfig {
		return nil
	}

	return config
}
