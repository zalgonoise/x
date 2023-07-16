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
	"github.com/zalgonoise/x/audio/fft"
	"github.com/zalgonoise/x/audio/fft/window"
	"github.com/zalgonoise/x/audio/wav"
)

type Reporter interface {
	SetPeakValue(data float64) (err error)
	SetPeakValues(data []float64) (err error)
	SetPeakFreq(frequency int, magnitude float64) (err error)
	io.Closer
}

type Stream struct {
	cfg    *config.Config
	logger logx.Logger

	proc   func([]float64) error
	out    Reporter
	stream *wav.Stream
}

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

func (s *Stream) Close() error {
	s.logger.Info("closing stream")

	return s.out.Close()
}

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

	switch cfg.Mode {
	case config.Monitor:
		s.proc = func(data []float64) error {
			var max float64

			for i := range data {
				if data[i] > max {
					max = data[i]
				}
			}

			return s.out.SetPeakValue(max)
		}
	case config.Analyze:
		bs := fft.Block1024
		windowBlock := window.New(window.Blackman, int(bs))

		s.proc = func(data []float64) error {
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

				return s.out.SetPeakFreq(freq, max)
			}

			return nil
		}
	}

	return s, nil
}
