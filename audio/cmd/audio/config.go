package main

import (
	"flag"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/zalgonoise/x/errs"
)

const (
	errDomain = errs.Domain("x/audio")

	ErrEmpty   = errs.Kind("empty")
	ErrInvalid = errs.Kind("invalid")

	ErrInput      = errs.Entity("input")
	ErrHTTPURL    = errs.Entity("HTTP or HTTPS URL")
	ErrHost       = errs.Entity("host")
	ErrPort       = errs.Entity("port")
	ErrOutputType = errs.Entity("output type")
	ErrMode       = errs.Entity("mode")
	ErrCompactor  = errs.Entity("compactor")
)

var (
	ErrEmptyInput        = errs.WithDomain(errDomain, ErrEmpty, ErrInput)
	ErrInvalidHTTPURL    = errs.WithDomain(errDomain, ErrInvalid, ErrHTTPURL)
	ErrInvalidHost       = errs.WithDomain(errDomain, ErrInvalid, ErrHost)
	ErrInvalidPort       = errs.WithDomain(errDomain, ErrInvalid, ErrPort)
	ErrInvalidOutputType = errs.WithDomain(errDomain, ErrInvalid, ErrOutputType)
	ErrInvalidMode       = errs.WithDomain(errDomain, ErrInvalid, ErrMode)
	ErrInvalidCompactor  = errs.WithDomain(errDomain, ErrInvalid, ErrCompactor)
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
	defaultPort           = 13088

	urlSchemeHTTP  = "http"
	urlSchemeHTTPS = "https"

	outputToLog        = "log"
	outputToLogger     = "logger"
	outputToProm       = "prom"
	outputToPrometheus = "prometheus"

	modePeaks    = "peaks"
	modeSpectrum = "spectrum"
	modeCombined = "combined"

	batchMax     = "max"
	batchMaximum = "maximum"
)

//nolint:gochecknoglobals // immutable object used in comparisons when creating Config
var emptyConfig = Config{}

func DefaultConfig() *Config {
	return &Config{
		OutputType: outputToLogger,
		BucketSize: defaultBucketSize,
		Mode:       modeCombined,
		Duration:   defaultDuration,
	}
}

type Config struct {
	// BufferSize defines the audio stream's buffer size in bytes.
	BufferSize int `envconfig:"X_AUDIO_BUFFER_SIZE"`
	// BufferDur defines the audio stream's buffer size as a duration.
	BufferDur time.Duration `envconfig:"X_AUDIO_BUFFER_DURATION"`
	// BufferRatio defines the audio stream's buffer size as a ratio to one second (1.0 is one second).
	BufferRatio float64 `envconfig:"X_AUDIO_BUFFER_RATIO"`

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

	StorageURI string `envconfig:"X_AUDIO_SQLITE_URI"`
}

func NewConfig() (*Config, error) {
	config := merge(
		merge(DefaultConfig(), newEnvConfig()),
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
	bufferSize := flag.Int("buffer-size", 0, "defines the audio stream's buffer size in bytes")
	bufferDur := flag.Duration("buffer-dur", 0, "defines the audio stream's buffer size as a duration")
	bufferRatio := flag.Float64("buffer-ratio", 0,
		"defines the audio stream's buffer size as a ratio to one second (1.0 is one second)")

	input := flag.String("input", "", "")

	outputType := flag.String("output-type", "", "")
	output := flag.String("output", "", "")

	mode := flag.String("mode", "", "")

	bucketSize := flag.Int("bucket-size", 0, "")
	batch := flag.Bool("batch", false, "")
	batchSize := flag.Int("batch-size", 0, "")
	batchFrequency := flag.Duration("batch-freq", 0, "")
	batchCompactor := flag.String("batch-compactor", "", "")

	duration := flag.Duration("dur", 0, "")
	exitCode := flag.Int("exit", 0, "")

	storageURI := flag.String("db.uri", "", "path to the SQLite instance to persist audio data")

	flag.Parse()

	config := &Config{
		BufferSize:     *bufferSize,
		BufferDur:      *bufferDur,
		BufferRatio:    *bufferRatio,
		Input:          *input,
		OutputType:     *outputType,
		Output:         *output,
		Mode:           *mode,
		BucketSize:     *bucketSize,
		Batch:          *batch,
		BatchSize:      *batchSize,
		BatchFrequency: *batchFrequency,
		BatchCompactor: *batchCompactor,
		Duration:       *duration,
		ExitCode:       *exitCode,
		StorageURI:     *storageURI,
	}

	if *config == emptyConfig {
		return nil
	}

	return config
}

func merge(base, input *Config) *Config {
	switch {
	case base == nil && input == nil:
		return DefaultConfig()
	case input == nil:
		return base
	case base == nil:
		return input
	}

	if input.BufferSize != 0 {
		base.BufferSize = input.BufferSize
	}

	if input.BufferDur != 0 {
		base.BufferDur = input.BufferDur
	}

	if input.BufferRatio != 0 {
		base.BufferRatio = input.BufferRatio
	}

	if input.Input != "" {
		base.Input = input.Input
	}

	if input.OutputType != "" {
		base.OutputType = input.OutputType
	}

	if input.Output != "" {
		base.Output = input.Output
	}

	if input.Mode != "" {
		base.Mode = input.Mode
	}

	if input.BucketSize > 0 {
		base.BucketSize = input.BucketSize
	}

	if input.Batch {
		base.Batch = input.Batch
	}

	if input.BatchSize > 0 {
		base.BatchSize = input.BatchSize
	}

	if input.BatchFrequency > 0 {
		base.BatchFrequency = input.BatchFrequency
	}

	if input.BatchCompactor != "" {
		base.BatchCompactor = input.BatchCompactor
	}

	if input.Duration > 0 {
		base.Duration = input.Duration
	}

	if input.ExitCode > 0 {
		base.ExitCode = input.ExitCode
	}

	if input.StorageURI != "" {
		base.StorageURI = input.StorageURI
	}

	return base
}

//nolint:cyclop // validate needs to perform several checks against the input *Config
func validate(config *Config) error {
	// validate input
	if config.Input == "" {
		return ErrEmptyInput
	}

	// currently there is only one (HTTP) consumer. We can try to
	// parse input as URL and default to OS Path if failed
	parsedURL, err := url.Parse(config.Input)
	if err != nil {
		return err
	}

	if parsedURL.Scheme != urlSchemeHTTP && parsedURL.Scheme != urlSchemeHTTPS {
		return fmt.Errorf("%w: %s", ErrInvalidHTTPURL, config.Input)
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("%w: %s", ErrInvalidHost, config.Input)
	}

	// validate output
	switch strings.ToLower(config.OutputType) {
	case outputToLogger, outputToLog:

	case outputToProm, outputToPrometheus:
		if _, err = getPort(config.Output); err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidPort, err)
		}
	default:
		return fmt.Errorf("%w: %s", ErrInvalidOutputType, config.OutputType)
	}

	// validate mode
	switch config.Mode {
	case modePeaks, modeSpectrum, modeCombined, "":
		// OK state
	default:
		return fmt.Errorf("%w: %s", ErrInvalidMode, config.Mode)
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
		case batchMax, batchMaximum, "":
		// OK state
		default:
			return fmt.Errorf("%w: %s", ErrInvalidCompactor, config.BatchCompactor)
		}
	}

	// validate runtime duration
	if config.Duration < minDuration || config.Duration > maxDuration {
		config.Duration = defaultDuration
	}

	return nil
}
