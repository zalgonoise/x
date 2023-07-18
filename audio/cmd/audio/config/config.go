package config

import (
	"time"
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
	// ExitCode forces a custom exit code on the processor when done or errored
	ExitCode int
}
