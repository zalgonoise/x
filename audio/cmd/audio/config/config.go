package config

import "time"

const (
	defaultMode               = Combined
	defaultDuration           = 30 * time.Second
	defaultOutput             = ToLogger
	defaultNumSpectrumBuckets = 64
)

// OpMode enumerates valid operation modes
type OpMode string

const (
	// Monitor is an OpMode that collects audio peak values in a stream
	Monitor OpMode = "monitor"
	// Analyze is an OpMode that collects audio peak frequencies in a stream
	Analyze OpMode = "analyze"
	// Combined is an OpMode that collects both audio peaks and peak frequencies in a stream
	Combined OpMode = "combined"
)

// Output enumerates valid output types
type Output string

const (
	// ToLogger is an Output that emits the collected metadata through a logx.Logger
	ToLogger Output = "logger"
	// ToFile is an Output that emits the collected metadata by writing it to a file in the system
	ToFile Output = "file"
	// ToPrometheus is an Output that emits the collected metadata as prometheus metrics, by exposing a metrics server
	ToPrometheus Output = "prom"
)

// Config describes an audio stream processor configuration
type Config struct {
	// Mode sets the operation mode for the processor
	Mode OpMode
	// URL points to an HTTP audio stream source
	URL string
	// Duration delimits a stream's runtime duration
	Duration time.Duration
	// Output sets the type of Output for the processor
	Output Output
	// OutputPath describes the path (or URL) for the set Output if applicable
	OutputPath string
	// NumSpectrumBuckets defines the number of buckets to distribute frequencies on, when analyzing a signal's spectrum
	NumSpectrumBuckets int
	// ExitCode forces a custom exit code on the processor when done or errored
	ExitCode int
}

// NewConfig creates a new Config by reading the input flags to the application startup
//
// It returns a new Config and an error, which is a call to the Validate(Config) function
func NewConfig() (*Config, error) {
	env := FromEnv()
	flags := FromFlags()

	switch {
	case env == nil && flags == nil:
		return nil, ErrMissingConfig
	case env == nil:
		return flags, Validate(flags)
	case flags == nil:
		return env, Validate(env)
	default:
		merged := Merge(env, flags)
		return merged, Validate(merged)
	}
}

// WithDefaults creates a new Config like NewConfig, but applies default values
// to any fields that have them unset, where applicable
func WithDefaults() (*Config, error) {
	conf, err := NewConfig()
	if err == nil {
		return conf, nil
	}

	c := applyDefaults(conf)
	return c, Validate(c)
}

func applyDefaults(c *Config) *Config {
	if c == nil {
		c = new(Config)
	}

	if c.Mode == "" {
		c.Mode = defaultMode
	}

	if c.Duration == 0 {
		c.Duration = defaultDuration
	}

	if c.Output == "" {
		c.Output = defaultOutput
	}

	if c.NumSpectrumBuckets == 0 {
		c.NumSpectrumBuckets = defaultNumSpectrumBuckets
	}

	return c
}

// Merge combines two Config, setting the values from `extra` in `main` where `main` has them unset
func Merge(main, extra *Config) *Config {
	if main == nil {
		main = new(Config)
	}

	if extra == nil {
		return main
	}

	if main.Mode == "" {
		main.Mode = extra.Mode
	}

	if main.URL == "" {
		main.URL = extra.URL
	}

	if main.Duration == 0 {
		main.Duration = extra.Duration
	}

	if main.Output == "" {
		main.Output = extra.Output
	}

	if main.OutputPath == "" {
		main.OutputPath = extra.OutputPath
	}

	if main.NumSpectrumBuckets == 0 {
		main.NumSpectrumBuckets = extra.NumSpectrumBuckets
	}

	if main.ExitCode == 0 {
		main.ExitCode = extra.ExitCode
	}

	return main
}
