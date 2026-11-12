// Command audio analyses an audio stream (a URL, a file or stdin): it reports per-channel level and spectrum metrics as
// JSON lines or on a Prometheus endpoint, and can record the stream as WAV segments. Run audio -h for the flags; each
// one also reads an AUDIO_* environment variable.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"net"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	promexp "github.com/zalgonoise/x/audio/contrib/prometheus"
	"github.com/zalgonoise/x/audio/dsp/spectral"
	"github.com/zalgonoise/x/audio/sdk"
	"github.com/zalgonoise/x/audio/sdk/metric"
	"github.com/zalgonoise/x/audio/sdk/recorder"
	"github.com/zalgonoise/x/audio/sdk/sources"
)

// Exit codes.
const (
	exitOK      = 0
	exitFailure = 1
	exitUsage   = 2
)

const (
	shutdownTimeout   = 15 * time.Second
	readHeaderTimeout = 10 * time.Second
)

func main() {
	// the command owns process-wide signal handling: SIGINT and SIGTERM cancel the run
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	code := run(ctx, os.Args[1:], os.Getenv, os.Stdin, os.Stdout, os.Stderr)

	stop()
	os.Exit(code)
}

// run is the whole command, with its environment injected: it returns the exit code.
func run(ctx context.Context, args []string, getenv func(string) string, stdin io.Reader, stdout, stderr io.Writer,
) int {
	cfg, err := parseConfig(args, getenv, stderr)

	switch {
	case errors.Is(err, flag.ErrHelp):
		return exitOK
	case err != nil:
		_, _ = fmt.Fprintln(stderr, err)

		return exitUsage
	}

	logger := newLogger(&cfg, stderr)

	var closers closerStack

	defer func() {
		// the run context may be cancelled already (signal, duration): shut down with a fresh deadline
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
		defer cancel()

		closers.shutdown(shutdownCtx, logger)
	}()

	p, err := build(&cfg, logger, stdin, stdout, &closers)
	if err != nil {
		logger.ErrorContext(ctx, "setting up", slog.Any("error", err))

		return exitFailure
	}

	if cfg.duration > 0 {
		var cancel context.CancelFunc

		ctx, cancel = context.WithTimeout(ctx, cfg.duration)
		defer cancel()
	}

	logger.InfoContext(ctx, "running", slog.String("input", cfg.input), slog.Duration("duration", cfg.duration))

	err = runLoop(ctx, p, &cfg, logger)

	switch {
	case err == nil:
		logger.InfoContext(ctx, "input ended")
	case errors.Is(err, context.DeadlineExceeded) && cfg.duration > 0:
		logger.InfoContext(ctx, "run duration reached")
	case errors.Is(err, context.Canceled):
		logger.InfoContext(ctx, "interrupted")
	default:
		logger.ErrorContext(ctx, "pipeline failed", slog.Any("error", err))

		return exitFailure
	}

	return exitOK
}

// runLoop runs the pipeline. With -reconnect and an http(s) input it runs it again whenever the stream ends or fails,
// waiting between attempts (exponential backoff with jitter, reset by a run longer than the maximum wait), until ctx
// ends.
func runLoop(ctx context.Context, p *sdk.Pipeline, cfg *config, logger *slog.Logger) error {
	in, err := parseInput(cfg.input)
	if err != nil {
		return err
	}

	wait := cfg.minWait

	for {
		start := time.Now()
		err := p.Run(ctx)

		switch {
		case ctx.Err() != nil:
			// the run was stopped (signal, -duration): whatever error the stream produced while it ended is not a failure
			return context.Cause(ctx)
		case !cfg.reconnect || in.url == "" || errors.Is(err, sdk.ErrShutdown):
			return err
		}

		if time.Since(start) > cfg.maxWait {
			wait = cfg.minWait // the stream was healthy for a while: start over
		}

		half := wait / backoffFactor
		delay := half + rand.N(half+1) //nolint:gosec // jitter, not security
		logger.WarnContext(ctx, "stream ended; reconnecting", slog.Any("error", err), slog.Duration("in", delay))

		timer := time.NewTimer(delay)

		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()

			return context.Cause(ctx)
		}

		wait = min(wait*backoffFactor, cfg.maxWait)
	}
}

func newLogger(cfg *config, w io.Writer) *slog.Logger {
	opts := &slog.HandlerOptions{Level: cfg.logLevel}

	if cfg.logJSON {
		return slog.New(slog.NewJSONHandler(w, opts))
	}

	return slog.New(slog.NewTextHandler(w, opts))
}

// build assembles the pipeline, pushing what must be shut down onto closers in creation order.
func build(cfg *config, logger *slog.Logger, stdin io.Reader, stdout io.Writer, closers *closerStack,
) (*sdk.Pipeline, error) {
	src, err := newSource(cfg, stdin)
	if err != nil {
		return nil, err
	}

	onError := func(err error) { logger.Warn("export failed", slog.Any("error", err)) }

	var exporters []sdk.Exporter

	if cfg.metrics != metricsNone {
		meter, err := newMeter(cfg, logger, stdout, onError, closers)
		if err != nil {
			return nil, err
		}

		exporters = append(exporters, meter)
	}

	if cfg.recordDir != "" {
		rec, err := newRecorder(cfg, onError, closers)
		if err != nil {
			return nil, err
		}

		exporters = append(exporters, rec)
	}

	p, err := sdk.NewPipeline(src, exporters, sdk.WithLogger(logger), sdk.WithBufferDuration(cfg.buffer),
		sdk.WithErrorHandler(onError))
	if err != nil {
		return nil, err
	}

	// last in, first out: the pipeline shuts down first, flushing every exporter before their backends close
	closers.push("pipeline", p.Shutdown)

	return p, nil
}

func newSource(cfg *config, stdin io.Reader) (sdk.Source, error) {
	in, err := parseInput(cfg.input)

	switch {
	case err != nil:
		return nil, err
	case in.stdin:
		return sources.NewReader(stdin), nil
	case in.url != "":
		return sources.NewHTTP(in.url)
	default:
		return sources.NewFile(in.path), nil
	}
}

// newMeter returns the level and spectrum meter behind a queue. For a live http(s) stream the queue drops the oldest
// buffers if the meter falls behind (real time, CLAUDE.md §7 Q8); a file or stdin decodes faster than real time, so
// there it blocks rather than skip audio. Dropped buffers are logged at shutdown.
func newMeter(cfg *config, logger *slog.Logger, stdout io.Writer, onError sdk.ErrorHandler, closers *closerStack,
) (sdk.Exporter, error) {
	var reader metric.Reader

	switch cfg.metrics {
	case metricsPrometheus:
		manual := metric.NewManualReader()

		if err := servePrometheus(cfg.promAddr, manual, logger, closers); err != nil {
			return nil, err
		}

		reader = manual
	default:
		emitter, err := metric.NewSlogEmitter(slog.New(slog.NewJSONHandler(stdout, nil)), slog.LevelInfo)
		if err != nil {
			return nil, err
		}

		periodic, err := metric.NewPeriodicReader(emitter, metric.WithInterval(cfg.interval),
			metric.WithErrorHandler(onError))
		if err != nil {
			return nil, err
		}

		reader = periodic
	}

	instruments, err := newInstruments(cfg)
	if err != nil {
		return nil, err
	}

	meter, err := metric.NewMeter(reader, instruments...)
	if err != nil {
		return nil, err
	}

	in, err := parseInput(cfg.input)
	if err != nil {
		return nil, err
	}

	policy := sdk.DropOldest
	if in.url == "" {
		policy = sdk.Block
	}

	q, err := sdk.NewQueuedExporter(meter, sdk.WithQueueSize(cfg.queue), sdk.WithPolicy(policy),
		sdk.WithQueueErrorHandler(onError))
	if err != nil {
		return nil, err
	}

	// pushed before the pipeline, so it runs after the pipeline (and this queue) shut down
	closers.push("metrics queue", func(ctx context.Context) error {
		if n := q.Dropped(); n > 0 {
			logger.WarnContext(ctx, "the meter fell behind the stream: buffers were not measured",
				slog.Uint64("dropped", n), slog.Int("queue", cfg.queue))
		}

		return nil
	})

	return q, nil
}

// Spectrum bands span the audible range.
const (
	lowestBand  = 20    // Hz
	highestBand = 20000 // Hz
)

// newInstruments returns the level instruments, plus the spectral ones that -metrics.spectrum asks for.
func newInstruments(cfg *config) ([]metric.Instrument, error) {
	instruments := metric.Levels()

	opts := []metric.InstrumentOption{metric.WithFrameSize(cfg.specSize)}

	switch cfg.spectrum {
	case spectrumNone:
		return instruments, nil
	case spectrumOctave, spectrumThirdOctave:
		fraction := spectral.ThirdOctave
		if cfg.spectrum == spectrumOctave {
			fraction = spectral.Octave
		}

		bands, err := spectral.OctaveBands(fraction, lowestBand, highestBand)
		if err != nil {
			return nil, err
		}

		opts = append(opts, metric.WithBands(bands))
	}

	spectra, err := metric.Spectra(opts...)
	if err != nil {
		return nil, err
	}

	return append(instruments, spectra...), nil
}

// servePrometheus serves the meter's measurements on addr/metrics until shutdown.
func servePrometheus(addr string, reader *metric.ManualReader, logger *slog.Logger, closers *closerStack) error {
	registry := prometheus.NewRegistry()

	for _, c := range []prometheus.Collector{
		promexp.NewCollector(reader),
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	} {
		if err := registry.Register(c); err != nil {
			return err
		}
	}

	ln, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", addr)
	if err != nil {
		return fmt.Errorf("audio: prometheus endpoint: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	srv := &http.Server{Handler: mux, ReadHeaderTimeout: readHeaderTimeout}

	go func() {
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("prometheus endpoint stopped", slog.Any("error", err))
		}
	}()

	logger.Info("prometheus endpoint listening", slog.String("addr", ln.Addr().String()))
	closers.push("prometheus endpoint", srv.Shutdown)

	return nil
}

// newRecorder returns the segment recorder, behind a queue that blocks rather than lose audio.
func newRecorder(cfg *config, onError sdk.ErrorHandler, closers *closerStack) (sdk.Exporter, error) {
	store, err := recorder.NewDir(cfg.recordDir)
	if err != nil {
		return nil, err
	}

	closers.push("record directory", func(context.Context) error { return store.Close() })

	rec, err := recorder.New(store, recorder.WithSegmentDuration(cfg.segment))
	if err != nil {
		return nil, err
	}

	return sdk.NewQueuedExporter(rec, sdk.WithQueueSize(cfg.queue), sdk.WithPolicy(sdk.Block),
		sdk.WithQueueErrorHandler(onError))
}

// closerStack shuts components down in the reverse order they were pushed.
type closerStack []namedCloser

type namedCloser struct {
	name     string
	shutdown func(ctx context.Context) error
}

func (s *closerStack) push(name string, fn func(ctx context.Context) error) {
	*s = append(*s, namedCloser{name: name, shutdown: fn})
}

func (s closerStack) shutdown(ctx context.Context, logger *slog.Logger) {
	for _, c := range slices.Backward(s) {
		if err := c.shutdown(ctx); err != nil {
			logger.WarnContext(ctx, "shutting down", slog.String("component", c.name), slog.Any("error", err))
		}
	}
}
