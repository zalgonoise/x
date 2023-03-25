package stream

import (
	"fmt"
	"time"

	"github.com/zalgonoise/x/ptr"
)

type err string

func (e err) Error() string { return (string)(e) }

const (
	ErrEmptyURL         err = "gsp/conf: empty URL"
	ErrModeNotSupported err = "gsp/conf: operation is not supported"
	ErrModeUnset        err = "gsp/conf: operation is undefined"
	ErrRecTimeUnset     err = "gsp/conf: recording time is undefined"
	ErrDurationUnset    err = "gsp/conf: runtime duration is undefined"
	ErrInvalidRatio     err = "gsp/conf: buffer size ratio cannot be zero or below"
	ErrEmptyDirectory   err = "gsp/conf: recording operation without a target file path"
	ErrEmptyThreshold   err = "gsp/conf: peak threshold is unset"
	ErrShortDuration    err = "gsp/conf: runtime duration is shorter than recording time"
)

const (
	defaultRecTime time.Duration = time.Second * 30
)

// Default is an initialized Config with sane default values
var Default = Config{
	Mode:       Monitor,
	Dur:        ptr.To(time.Second * 30),
	BufferSize: 0.5,
}

// Config describes the app's configuration
type Config struct {
	URL        string         // URL points to the target host exposing the audio stream
	Mode       Mode           // Mode is an enumeration selection of the operation mode of the app
	Dur        *time.Duration // Dur delimits the app's runtime duration
	RecTime    *time.Duration // RecTime delimits a recording's duration
	BufferSize float64        // BufferSize is a ratio for the ring buffer's size (1.0 is 1 second; 0.5 is 500ms; etc)
	Peak       *int           // Peak is the peak PCM integer value that will trigger recording the stream
	Dir        *string        // Dir is the output directory (and filename prefix) where the recording(s) should be stored
	Prom       bool           // Prom is a boolean to set the output as a Prometheus /metrics HTTP endpoint; instead of os.Stdout
	Port       int            // Port defines an override to the Prometheus metrics port if defined
	ExitCode   int            // ExitCode defines an override to the exit code of the application
}

// Merge combines Config `c` with Config `input`, returning a merged version
// of the two
//
// All set elements in Config `input` will be applied to Config `c`, and the unset elements
// will be ignored (keeps Config `c`'s data)
func (c *Config) Merge(input *Config) *Config {
	if input.URL != "" {
		c.URL = input.URL
	}
	if input.Mode != Unset {
		c.Mode = input.Mode
	}
	if input.Dur != nil {
		c.Dur = input.Dur
	}
	if input.RecTime != nil {
		c.RecTime = input.RecTime
	}
	if input.BufferSize > 0 {
		c.BufferSize = input.BufferSize
	}
	if input.Peak != nil {
		c.Peak = input.Peak
	}
	if input.Dir != nil {
		c.Dir = input.Dir
	}
	if input.Prom {
		c.Prom = input.Prom
	}
	if input.Port > 0 {
		c.Port = input.Port
	}
	if input.ExitCode > 0 {
		c.ExitCode = input.ExitCode
	}
	return c
}

// Apply implements the Option interface
//
// It allows applying new options on top of an already existing config
func (c *Config) Apply(opts ...Option) *Config {
	for _, opt := range opts {
		if opt != nil {
			opt.Apply(c)
		}
	}
	return c
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
			return ErrRecTimeUnset
		}
		if c.Dir == nil {
			return ErrEmptyDirectory
		}
	case Filter:
		if c.RecTime == nil {
			return ErrRecTimeUnset
		}
		if c.Peak == nil || *c.Peak == 0 {
			return ErrEmptyThreshold
		}
		if c.Dir == nil {
			return ErrEmptyDirectory
		}
	case Unset:
		return ErrModeUnset
	default:
		return fmt.Errorf("%w: invalid mode: %d", ErrModeNotSupported, c.Mode)
	}

	if c.Dur == nil {
		return ErrDurationUnset
	}

	if c.RecTime != nil && *c.Dur < *c.RecTime {
		return ErrShortDuration
	}

	if c.BufferSize <= 0 {
		return ErrInvalidRatio
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
