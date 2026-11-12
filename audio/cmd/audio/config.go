package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/zalgonoise/x/audio/dsp/spectral"
	"github.com/zalgonoise/x/audio/sdk"
	"github.com/zalgonoise/x/audio/sdk/recorder"
)

// Errors returned when parsing the configuration; all of them are usage errors.
var (
	ErrMissingInput  = errors.New("audio: -input is required")
	ErrInvalidInput  = errors.New("audio: invalid -input")
	ErrInvalidValue  = errors.New("audio: invalid value")
	ErrInvalidEnv    = errors.New("audio: invalid environment variable")
	ErrUnexpectedArg = errors.New("audio: unexpected argument")
)

// Destinations of the metrics (-metrics).
const (
	metricsLog        = "log"
	metricsPrometheus = "prometheus"
	metricsNone       = "none"
)

// Spectrum layouts (-metrics.spectrum): fractional-octave bands, every FFT bin, or no spectrum.
const (
	spectrumNone        = "none"
	spectrumOctave      = "octave"
	spectrumThirdOctave = "third-octave"
	spectrumFull        = "full"
)

const defaultSpectrumSize = 4096 // 93 ms at 44.1 kHz, bins 10.8 Hz apart: enough for the 31.5 Hz third octave

// Reconnect waits: the first wait doubles after each failed attempt, up to the maximum.
const (
	defaultMinWait = time.Second
	defaultMaxWait = 30 * time.Second
	backoffFactor  = 2
)

// envPrefix prefixes the environment variable of every flag: -metrics.interval reads AUDIO_METRICS_INTERVAL.
const envPrefix = "AUDIO_"

// config is the command's configuration. Every flag falls back to its environment variable when it is not on the
// command line.
type config struct {
	input     string
	duration  time.Duration
	buffer    time.Duration
	metrics   string
	interval  time.Duration
	spectrum  string
	specSize  int
	promAddr  string
	recordDir string
	segment   time.Duration
	queue     int
	reconnect bool
	minWait   time.Duration
	maxWait   time.Duration
	logLevel  slog.Level
	logJSON   bool
}

func envName(flagName string) string {
	return envPrefix + strings.ToUpper(strings.NewReplacer(".", "_", "-", "_").Replace(flagName))
}

// parseConfig parses args, filling flags not given on the command line from getenv. It returns flag.ErrHelp for -h.
func parseConfig(args []string, getenv func(string) string, stderr io.Writer) (config, error) {
	fs := flag.NewFlagSet("audio", flag.ContinueOnError)
	fs.SetOutput(stderr)

	usage := func(name, text string) string { return text + " (env " + envName(name) + ")" }

	var (
		c         config
		logLevel  string
		logFormat string
	)

	fs.StringVar(&c.input, "input", "", usage("input",
		"audio to analyse: an http(s) URL, a file path, a file:// URL, or - for stdin"))
	fs.DurationVar(&c.duration, "duration", 0, usage("duration",
		"stop after this long; 0 runs until the input ends or a signal arrives"))
	fs.DurationVar(&c.buffer, "buffer", sdk.DefaultBufferDuration, usage("buffer", "audio carried by each buffer"))
	fs.StringVar(&c.metrics, "metrics", metricsLog, usage("metrics",
		"where per-channel level and spectrum metrics go: log (JSON lines on stdout), prometheus, or none"))
	fs.DurationVar(&c.interval, "metrics.interval", time.Second, usage("metrics.interval",
		"how often logged metrics are emitted"))
	fs.StringVar(&c.spectrum, "metrics.spectrum", spectrumThirdOctave, usage("metrics.spectrum",
		"per-channel spectrum (peak hold and mean): third-octave or octave bands from 20 Hz to 20 kHz, full (every FFT "+
			"bin; too many series for prometheus), or none"))
	fs.IntVar(&c.specSize, "metrics.spectrum.size", defaultSpectrumSize, usage("metrics.spectrum.size",
		"samples per analysed frame: a power of two; larger frames resolve lower bands"))
	fs.StringVar(&c.promAddr, "prometheus.addr", ":13088", usage("prometheus.addr",
		"listen address of the /metrics endpoint, with -metrics prometheus"))
	fs.StringVar(&c.recordDir, "record.dir", "", usage("record.dir",
		"record the audio as WAV segments in this directory (off when empty)"))
	fs.DurationVar(&c.segment, "record.segment", recorder.DefaultSegmentDuration, usage("record.segment",
		"duration of each recorded segment"))
	fs.IntVar(&c.queue, "queue", sdk.DefaultQueueSize, usage("queue", "buffers queued in front of each exporter"))
	fs.BoolVar(&c.reconnect, "reconnect", false, usage("reconnect",
		"for an http(s) input, reconnect when the stream ends or fails, until -duration or a signal"))
	fs.DurationVar(&c.minWait, "reconnect.min-wait", defaultMinWait, usage("reconnect.min-wait",
		"first wait before reconnecting; it doubles after each failed attempt"))
	fs.DurationVar(&c.maxWait, "reconnect.max-wait", defaultMaxWait, usage("reconnect.max-wait",
		"longest wait before reconnecting; a run longer than this resets the wait"))
	fs.StringVar(&logLevel, "log.level", "info", usage("log.level", "debug, info, warn or error"))
	fs.StringVar(&logFormat, "log.format", "text", usage("log.format", "text or json"))

	if err := fs.Parse(args); err != nil {
		return config{}, err
	}

	if fs.NArg() > 0 {
		return config{}, fmt.Errorf("%w: %q", ErrUnexpectedArg, fs.Arg(0))
	}

	if err := applyEnv(fs, getenv); err != nil {
		return config{}, err
	}

	if err := c.logLevel.UnmarshalText([]byte(logLevel)); err != nil {
		return config{}, fmt.Errorf("%w: -log.level %q", ErrInvalidValue, logLevel)
	}

	switch logFormat {
	case "text":
	case "json":
		c.logJSON = true
	default:
		return config{}, fmt.Errorf("%w: -log.format %q", ErrInvalidValue, logFormat)
	}

	if err := c.validate(); err != nil {
		return config{}, err
	}

	return c, nil
}

// applyEnv sets every flag that was not given on the command line from its environment variable, if set.
func applyEnv(fs *flag.FlagSet, getenv func(string) string) error {
	given := map[string]bool{}

	fs.Visit(func(f *flag.Flag) { given[f.Name] = true })

	var errs []error

	fs.VisitAll(func(f *flag.Flag) {
		if given[f.Name] {
			return
		}

		name := envName(f.Name)
		if v := getenv(name); v != "" {
			if err := fs.Set(f.Name, v); err != nil {
				errs = append(errs, fmt.Errorf("%w: %s=%q: %w", ErrInvalidEnv, name, v, err))
			}
		}
	})

	return errors.Join(errs...)
}

func (c *config) validate() error {
	switch {
	case c.input == "":
		return ErrMissingInput
	case c.duration < 0:
		return fmt.Errorf("%w: -duration %s", ErrInvalidValue, c.duration)
	case c.buffer <= 0:
		return fmt.Errorf("%w: -buffer %s", ErrInvalidValue, c.buffer)
	case c.interval <= 0:
		return fmt.Errorf("%w: -metrics.interval %s", ErrInvalidValue, c.interval)
	case c.segment <= 0:
		return fmt.Errorf("%w: -record.segment %s", ErrInvalidValue, c.segment)
	case c.queue < 1:
		return fmt.Errorf("%w: -queue %d", ErrInvalidValue, c.queue)
	case c.minWait <= 0 || c.maxWait < c.minWait:
		return fmt.Errorf("%w: -reconnect.min-wait %s, -reconnect.max-wait %s", ErrInvalidValue, c.minWait, c.maxWait)
	}

	switch c.metrics {
	case metricsLog, metricsPrometheus, metricsNone:
	default:
		return fmt.Errorf("%w: -metrics %q (want log, prometheus or none)", ErrInvalidValue, c.metrics)
	}

	switch c.spectrum {
	case spectrumNone, spectrumOctave, spectrumThirdOctave, spectrumFull:
	default:
		return fmt.Errorf("%w: -metrics.spectrum %q (want third-octave, octave, full or none)", ErrInvalidValue,
			c.spectrum)
	}

	if n := c.specSize; n < spectral.MinSize || n > spectral.MaxSize || n&(n-1) != 0 {
		return fmt.Errorf("%w: -metrics.spectrum.size %d (want a power of two in [%d, %d])", ErrInvalidValue, n,
			spectral.MinSize, spectral.MaxSize)
	}

	_, err := parseInput(c.input)

	return err
}

// input is where the audio comes from: exactly one of the fields is set.
type input struct {
	stdin bool
	url   string
	path  string
}

// parseInput classifies -input: "-" is stdin, http(s) URLs are fetched, file:// URLs and anything else are local paths,
// which must exist.
func parseInput(raw string) (input, error) {
	if raw == "-" {
		return input{stdin: true}, nil
	}

	if u, err := url.Parse(raw); err == nil {
		switch u.Scheme {
		case "http", "https":
			if u.Host == "" {
				return input{}, fmt.Errorf("%w: %q has no host", ErrInvalidInput, raw)
			}

			return input{url: raw}, nil
		case "file":
			raw = u.Path
		case "":
		default:
			return input{}, fmt.Errorf("%w: unsupported scheme %q", ErrInvalidInput, u.Scheme)
		}
	}

	info, err := os.Stat(raw) //nolint:gosec // the path comes from the operator, who may read any file they choose

	switch {
	case err != nil:
		return input{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	case info.IsDir():
		return input{}, fmt.Errorf("%w: %q is a directory", ErrInvalidInput, raw)
	}

	return input{path: raw}, nil
}
