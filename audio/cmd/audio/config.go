package main

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
)

const (
	minBucketSize         = 8
	defaultBucketSize     = 64
	minBatchSize          = 8
	defaultBatchSize      = 64
	minBatchFrequency     = 20 * time.Millisecond
	maxBatchFrequency     = 5 * time.Minute
	defaultBatchFrequency = 200 * time.Millisecond
	minDuration           = 15 * time.Second
	maxDuration           = 365 * 24 * time.Hour
	defaultDuration       = 4 * time.Hour
	defaultOutputType     = "logger"
	defaultMode           = "combined"
	defaultPort           = 13088
)

var (
	emptyConfig   = Config{}
	defaultConfig = Config{
		OutputType: defaultOutputType,
		BucketSize: defaultBucketSize,
		Mode:       defaultMode,
		Duration:   defaultDuration,
	}
)

type Config struct {
	// Input points to an audio stream source (URL)
	Input string `envconfig:"X_AUDIO_INPUT"`

	// OutputType sets the type of exporter for the processor
	OutputType string `envconfig:"X_AUDIO_OUTPUT_TYPE"`
	// Output describes the path (or URL) for the set Output if applicable
	Output string `envconfig:"X_AUDIO_OUTPUT"`

	// Mode sets the operation mode for the processor
	Mode string `envconfig:"X_AUDIO_MODE"`

	// BucketSize defines the number of frequencies in a bucket, when analyzing a signal's spectrum
	BucketSize int `envconfig:"X_AUDIO_BUCKET_SIZE"`
	// Batch defines if the collector batches the values or if it constantly registers them
	Batch bool `envconfig:"X_AUDIO_BATCH"`
	// BatchSize defines the size of each batching pass
	BatchSize int `envconfig:"X_AUDIO_BATCH_SIZE"`
	// BatchFrequency defines how often the batched values are flushed
	BatchFrequency time.Duration `envconfig:"X_AUDIO_BATCH_FREQUENCY"`
	// BatchCompactor defines the compactor (reduce) strategy when batching
	BatchCompactor string `envconfig:"X_AUDIO_BATCH_COMPACTOR"`

	// Duration delimits a stream's runtime duration
	Duration time.Duration `envconfig:"X_AUDIO_DURATION"`
	// ExitCode forces a custom exit code on the processor when done or errored
	ExitCode int `envconfig:"X_AUDIO_EXIT_CODE"`
}

func NewConfig() (*Config, error) {
	config := merge(
		merge(&defaultConfig, newEnvConfig()),
		newFlagsConfig(),
	)

	return config, validate(config)
}

func newEnvConfig() *Config {
	config := new(Config)

	if err := envconfig.Process("", config); err != nil {
		return nil
	}

	if *config == emptyConfig {
		return nil
	}

	return config
}

func newFlagsConfig() *Config {
	input := flag.String("input", "", "")

	outputType := flag.String("output-type", "", "")
	output := flag.String("output", "", "")

	mode := flag.String("mode", "", "")

	bucketSize := flag.Int("bucket-size", 0, "")
	batch := flag.Bool("batch", false, "")
	batchSize := flag.Int("batch-size", 0, "")
	batchFrequencyString := flag.String("batch-freq", "", "")
	batchCompactor := flag.String("batch-compactor", "", "")

	durationString := flag.String("dur", "", "")
	exitCode := flag.Int("exit", 0, "")

	flag.Parse()

	batchFrequency, err := time.ParseDuration(*batchFrequencyString)
	if err != nil {
		return nil
	}

	duration, err := time.ParseDuration(*durationString)
	if err != nil {
		return nil
	}

	config := &Config{
		Input:          *input,
		OutputType:     *outputType,
		Output:         *output,
		Mode:           *mode,
		BucketSize:     *bucketSize,
		Batch:          *batch,
		BatchSize:      *batchSize,
		BatchFrequency: batchFrequency,
		BatchCompactor: *batchCompactor,
		Duration:       duration,
		ExitCode:       *exitCode,
	}

	if *config == emptyConfig {
		return nil
	}

	return config
}

func merge(base, new *Config) *Config {
	switch {
	case base == nil && new == nil:
		return &defaultConfig
	case new == nil:
		return base
	case base == nil:
		return new
	}

	if new.Input != "" {
		base.Input = new.Input
	}

	if new.OutputType != "" {
		base.OutputType = new.OutputType
	}

	if new.Output != "" {
		base.Output = new.Output
	}

	if new.Mode != "" {
		base.Mode = new.Mode
	}

	if new.BucketSize > 0 {
		base.BucketSize = new.BucketSize
	}

	if new.Batch {
		base.Batch = new.Batch
	}

	if new.BatchSize > 0 {
		base.BatchSize = new.BatchSize
	}

	if new.BatchFrequency > 0 {
		base.BatchFrequency = new.BatchFrequency
	}

	if new.BatchCompactor != "" {
		base.BatchCompactor = new.BatchCompactor
	}

	if new.Duration > 0 {
		base.Duration = new.Duration
	}

	if new.ExitCode > 0 {
		base.ExitCode = new.ExitCode
	}

	return base
}

func validate(config *Config) error {
	// validate input
	if config.Input == "" {
		return errors.New("input cannot be empty")
	}

	// currently there is only one (HTTP) consumer. We can try to
	// parse input as URL and default to OS Path if failed
	parsedURL, err := url.Parse(config.Input)
	if err != nil {
		return err
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("input must be a valid HTTP or HTTPS URL: %s", config.Input)
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("input must be a valid host: %s", config.Input)
	}

	// validate output
	switch strings.ToLower(config.OutputType) {
	case "logger", "log":

	case "prom", "prometheus":
		if config.Output != "" {
			split := strings.Split(config.Output, ":")

			switch len(split) {
			case 0:
				// apply defaults
			case 1:
				_, err := strconv.Atoi(split[0])
				if err != nil {
					return fmt.Errorf("output must be a valid port number: %v", err)
				}

			case 2:
				// assume "localhost" "{port}"
				_, err := strconv.Atoi(split[1])
				if err != nil {
					return fmt.Errorf("output must be a valid port number: %v", err)
				}
			default:
				return fmt.Errorf(
					"invalid output address, should be either a port or 'localhost:{port}': %s", config.Output,
				)
			}
		}
	default:
		return fmt.Errorf("invalid output type: %s", config.OutputType)
	}

	// validate mode
	switch config.Mode {
	case "peaks", "spectrum", "combined", "":
		// OK state
	default:
		return fmt.Errorf("invalid mode: %s", config.Mode)
	}

	// validate spectrum bucket size
	if config.BucketSize < minBucketSize {
		config.BucketSize = defaultBucketSize
	}

	// validate batching config
	if config.Batch {
		if config.BatchSize < minBatchSize {
			config.BatchSize = defaultBatchSize
		}

		if config.BatchFrequency < minBatchFrequency || config.BatchFrequency > maxBatchFrequency {
			config.BatchFrequency = defaultBatchFrequency
		}

		switch config.BatchCompactor {
		case "max", "maximum", "":
		// OK state
		default:
			return fmt.Errorf("invalid compactor: %s", config.BatchCompactor)
		}
	}

	// validate runtime duration
	if config.Duration < minDuration || config.Duration > maxDuration {
		config.Duration = defaultDuration
	}

	return nil
}