package stream

import (
	"context"
	"errors"
	"io"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/zalgonoise/logx"
	"github.com/zalgonoise/logx/handlers/texth"

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
	// SetPeakFreq registers the int `frequency` value as an audio peak frequency
	SetPeakFreq(frequency int, magnitude float64) (err error)
	// Closer is an io.Closer, that is used to gracefully stop the Reporter
	io.Closer
}

// Stream describes an instance of an audio stream processor
type Stream struct {
	cfg    *config.Config
	logger logx.Logger

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

		logger := logx.New(texth.New(io.MultiWriter(f, os.Stderr)))
		s.out = NewLogWriter("", logger)
		s.logger = logger
	case config.ToPrometheus:
		port, err := strconv.Atoi(cfg.OutputPath)
		if err != nil {
			port = 0
		}

		prom, err := NewPromWriter(port)
		if err != nil {
			return nil, err
		}

		s.logger = logx.New(texth.New(os.Stderr))
		s.out = prom
	default:
		logger := logx.New(texth.New(os.Stderr))
		s.out = NewLogWriter("", logger)
		s.logger = logger
	}

	procFns := make([]func([]float64) error, 0)

	switch cfg.Mode {
	case config.Monitor:
		procFns = append(procFns, newMontiorFunc(s))
	case config.Analyze:
		procFns = append(procFns, newAnalyzeFunc(s))
	case config.Combined:
		procFns = append(procFns, newMontiorFunc(s), newAnalyzeFunc(s))
	default:
		return nil, ErrInvalidMode
	}

	s.proc = wav.MultiProc(false, procFns...)

	return s, nil
}

func newMontiorFunc(s *Stream) func([]float64) error {
	return func(data []float64) error {
		var max float64

		for i := range data {
			if data[i] > max {
				max = data[i]
			}
		}

		return s.out.SetPeakValue(max)
	}
}

func newAnalyzeFunc(s *Stream) func([]float64) error {
	bs := fft.Block1024
	windowBlock := window.New(window.Blackman, int(bs))

	return func(data []float64) error {
		for i := 0; i+int(bs) < len(data); i += int(bs) {
			spectrum := fft.Apply(
				int(s.stream.Wav.Header.SampleRate),
				data[i:i+int(bs)],
				windowBlock,
			)

			var max float64
			var freq int

			for idx := range spectrum {
				if max < spectrum[idx].Mag {
					max = spectrum[idx].Mag
					freq = spectrum[idx].Freq
				}
			}

			if err := s.out.SetPeakFreq(freq, max); err != nil {
				return err
			}
		}

		return nil
	}
}
