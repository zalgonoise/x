package stream

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/zalgonoise/x/audio/cmd/audio/config"
	"github.com/zalgonoise/x/audio/cmd/audio/http"
	"github.com/zalgonoise/x/audio/errs"
	"github.com/zalgonoise/x/audio/fft"
	"github.com/zalgonoise/x/audio/fft/window"
	"github.com/zalgonoise/x/audio/wav"
)

const (
	streamDomain = errs.Domain("audio/stream")

	ErrInvalid = errs.Kind("invalid")

	ErrMode = errs.Entity("operation mode")
)

var ErrInvalidMode = errs.New(streamDomain, ErrInvalid, ErrMode)

// Reporter describes the actions that a config.Output should expose
type Reporter interface {
	// SetPeakValue registers the float64 `data` value as an audio peak
	SetPeakValue(data float64) (err error)
	// ObserveFrequencies keeps track of changes in the registered frequencies
	ObserveFrequencies([]fft.FrequencyPower) (err error)
	// Closer is an io.Closer, that is used to gracefully stop the Reporter
	io.Closer
}

// Stream describes an instance of an audio stream processor
type Stream struct {
	cfg    *config.Config
	logger *slog.Logger

	proc   func([]float64) error
	out    Reporter
	stream *wav.Stream
}

// Run will consume the audio from the stream, returning an error if raised. Run is a blocking call.
func (s *Stream) Run(ctx context.Context) error {
	var (
		streamCtx = ctx
		done      context.CancelFunc
		errCh     = make(chan error)
		sigCh     = make(chan os.Signal, 1)
	)

	s.stream = wav.NewStream(nil, s.proc)

	signal.Notify(sigCh, os.Interrupt, os.Kill, syscall.SIGTERM)

	response, cancel, err := http.New(s.logger, s.cfg.URL, s.cfg.Duration)
	if err != nil {
		return err
	}

	defer cancel()

	if s.cfg.Duration > 0 {
		streamCtx, done = context.WithTimeout(ctx, s.cfg.Duration)

		defer done()
	}

	go s.stream.Stream(streamCtx, response.Body, errCh)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-sigCh:
			return ctx.Err()
		case err = <-errCh:
			if errors.Is(err, context.DeadlineExceeded) {
				return nil
			}

			return err
		}
	}
}

// Close implements the io.Closer interface, to gracefully stop the audio stream
func (s *Stream) Close() error {
	s.logger.Info("closing stream")

	return s.out.Close()
}

// New creates a Stream from the input Config, also returning an error if raised
func New(cfg *config.Config) (*Stream, error) {
	s := &Stream{
		cfg: cfg,
	}

	switch cfg.Output {
	case config.ToFile:
		f, err := os.OpenFile(cfg.OutputPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0664)
		if err != nil {
			return nil, err
		}

		logger := slog.New(slog.NewTextHandler(io.MultiWriter(f, os.Stderr), &slog.HandlerOptions{
			AddSource: true,
		}))
		s.out = NewLogWriter("", logger)
		s.logger = logger
	case config.ToPrometheus:
		prom, err := NewPromWriter(cfg.OutputPath)
		if err != nil {
			return nil, err
		}

		s.logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			AddSource: true,
		}))
		s.out = prom
	default:
		logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			AddSource: true,
		}))
		s.out = NewLogWriter("", logger)
		s.logger = logger
	}

	procFns := make([]func([]float64) error, 0)

	switch cfg.Mode {
	case config.Monitor:
		procFns = append(procFns, newMontiorFunc(s))
	case config.Analyze:
		procFns = append(procFns, newAnalyzeFunc(s, cfg.NumSpectrumBuckets))
	case config.Combined:
		procFns = append(procFns, newMontiorFunc(s), newAnalyzeFunc(s, cfg.NumSpectrumBuckets))
	default:
		return nil, ErrInvalidMode
	}

	s.proc = wav.MultiProc(false, procFns...)

	return s, nil
}

func newMontiorFunc(s *Stream) func([]float64) error {
	return func(data []float64) error {
		var maximum float64

		for i := range data {
			if data[i] > maximum {
				maximum = data[i]
			}
		}

		return s.out.SetPeakValue(maximum)
	}
}

func newAnalyzeFunc(s *Stream, blockSize int) func([]float64) error {
	bs := fft.NearestBlock(blockSize)
	windowBlock := window.New(window.Blackman, int(bs))

	return func(data []float64) error {
		for i := 0; i+int(bs) < len(data); i += int(bs) {
			spectrum := fft.Apply(
				int(s.stream.Wav.Header.SampleRate),
				data[i:i+int(bs)],
				windowBlock,
			)

			if err := s.out.ObserveFrequencies(spectrum); err != nil {
				return err
			}
		}

		return nil
	}
}
