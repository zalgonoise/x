package stream

import (
	"fmt"
	"time"

	"github.com/zalgonoise/x/audio/errs"
)

const (
	confDomain = errs.Domain("gps/conf")

	ErrEmpty        = errs.Kind("empty")
	ErrUnsupported  = errs.Kind("unsupported")
	ErrLessThanZero = errs.Kind("less than zero")
	ErrShort        = errs.Kind("short")

	ErrURL             = errs.Entity("URL")
	ErrMode            = errs.Entity("operation mode")
	ErrRecordingTime   = errs.Entity("recording time")
	ErrRuntimeDuration = errs.Entity("runtime duration")
	ErrBufferSizeRatio = errs.Entity("buffer size ratio")
	ErrBlockSize       = errs.Entity("block size")
	ErrFilePath        = errs.Entity("file path")
	ErrPeakThreshold   = errs.Entity("peak threshold")
)

var (
	ErrEmptyURL         = errs.New(confDomain, ErrEmpty, ErrURL)
	ErrUnsupportedMode  = errs.New(confDomain, ErrUnsupported, ErrMode)
	ErrUnsetMode        = errs.New(confDomain, ErrEmpty, ErrMode)
	ErrUnsetRecTime     = errs.New(confDomain, ErrEmpty, ErrRecordingTime)
	ErrUnsetDuration    = errs.New(confDomain, ErrEmpty, ErrRuntimeDuration)
	ErrInvalidRatio     = errs.New(confDomain, ErrLessThanZero, ErrBufferSizeRatio)
	ErrInvalidBlockSize = errs.New(confDomain, ErrLessThanZero, ErrBlockSize)
	ErrEmptyDirectory   = errs.New(confDomain, ErrEmpty, ErrFilePath)
	ErrEmptyThreshold   = errs.New(confDomain, ErrEmpty, ErrPeakThreshold)
	ErrShortDuration    = errs.New(confDomain, ErrShort, ErrRuntimeDuration)
)

var defaultRecTime = 30 * time.Second

// Default is an initialized Config with sane default values
var Default = Config{
	Mode:       Monitor,
	Dur:        &defaultRecTime,
	BufferSize: 0.5,
}

// Config describes the app's configuration
type Config struct {
	URL        string         // URL points to the target host exposing the audio stream
	Mode       Mode           // Mode is an enumeration selection of the operation mode of the app
	Dur        *time.Duration // Dur delimits the app's runtime duration
	RecTime    *time.Duration // RecTime delimits a recording's duration
	BufferSize float64        // BufferSize is a ratio for the ring buffer's size (1.0 is 1 second; 0.5 is 500ms; etc)
	BlockSize  int            // BlockSize is a specific size for the ring buffer, when a specific concrete value is needed
	Peak       []int          // Peak is the peak PCM integer value that will trigger recording the stream
	FloatPeaks []float64      // FloatPeaks is the floating-point value that will trigger recording the stream
	Dir        *string        // Dir is the output directory (and filename prefix) where the recording(s) should be stored
	Prom       bool           // Prom is a boolean to set the output as a Prometheus /metrics HTTP endpoint; instead of os.Stdout
	Port       int            // Port defines an override to the Prometheus metrics port if defined
	ExitCode   int            // ExitCode defines an override to the exit code of the application
}

// Apply implements the Option interface
//
// It allows applying new options on top of an already existing config
func (c *Config) Apply(input *Config) {
	if c.URL != "" {
		input.URL = c.URL
	}
	if c.Mode != Unset {
		input.Mode = c.Mode
	}
	if c.Dur != nil {
		input.Dur = c.Dur
	}
	if c.RecTime != nil {
		input.RecTime = c.RecTime
	}
	if c.BufferSize > 0 {
		input.BufferSize = c.BufferSize
	}
	if c.BlockSize > 0 {
		input.BlockSize = c.BlockSize
	}
	if len(c.Peak) > 0 {
		input.Peak = c.Peak
	}
	if c.Dir != nil {
		input.Dir = c.Dir
	}
	if c.Prom {
		input.Prom = c.Prom
	}
	if c.Port > 0 {
		input.Port = c.Port
	}
	if c.ExitCode > 0 {
		input.ExitCode = c.ExitCode
	}
}

// Validate verifies the elements in the Config `c` to ensure they are valid
//
// A corresponding error is returned if any invalid data is found
func (c *Config) Validate() error {
	if c.URL == "" {
		return ErrEmptyURL
	}

	switch c.Mode {
	case Monitor:
	case Record:
		if c.RecTime == nil {
			return ErrUnsetRecTime
		}
		if c.Dir == nil {
			return ErrEmptyDirectory
		}
	case Filter:
		if c.RecTime == nil {
			return ErrUnsetRecTime
		}
		if len(c.Peak) == 0 {
			return ErrEmptyThreshold
		}
		if c.Dir == nil {
			return ErrEmptyDirectory
		}
	case Analyze:
	case Unset:
		return ErrUnsetMode
	default:
		return fmt.Errorf("%w: %d", ErrUnsupportedMode, c.Mode)
	}

	if c.Dur == nil {
		return ErrUnsetDuration
	}

	if c.RecTime != nil && *c.Dur < *c.RecTime {
		return ErrShortDuration
	}

	if c.BufferSize <= 0 {
		return ErrInvalidRatio
	}

	if c.BlockSize < 0 {
		return ErrInvalidBlockSize
	}

	return nil
}

func newConfig(opts ...Option) *Config {
	conf := &Default

	for _, opt := range opts {
		if opt != nil {
			opt.Apply(conf)
		}
	}

	return conf
}

// Option describes setter types for a Config
//
// As new options / elements are added to the Config, new data structures can
// implement the Option interface to allow setting these options in the Config
type Option interface {
	// Apply sets the configuration on the input Config `c`
	Apply(c *Config)
}

// NewConfig initializes a new Config with Default settings, and then iterates through
// all input Option `opts` applying them to the Config, which is returned
// to the caller; alongside an error if raised
func NewConfig(opts ...Option) (*Config, error) {
	c := newConfig(opts...)
	return c, c.Validate()
}
