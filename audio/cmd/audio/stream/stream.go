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
	"github.com/zalgonoise/x/audio/wav"
)

type Reporter interface {
	SetPeakValue(data float64) (err error)
	SetPeakValues(data []float64) (err error)
	io.Closer
}

type Stream struct {
	cfg *config.Config

	proc func([]float64) error
	out  Reporter
}

func (s Stream) Run(ctx context.Context) error {
	var (
		w         = wav.NewStream(nil, s.proc)
		streamCtx = ctx
		done      context.CancelFunc
		errCh     = make(chan error)
		sigCh     = make(chan os.Signal, 1)
	)

	signal.Notify(sigCh, os.Interrupt, os.Kill, syscall.SIGTERM)

	response, cancel, err := http.New(s.cfg.URL, s.cfg.Duration)
	if err != nil {
		return err
	}

	defer cancel()

	if s.cfg.Duration > 0 {
		streamCtx, done = context.WithTimeout(ctx, s.cfg.Duration)
		defer done()
	}

	go w.Stream(streamCtx, response.Body, errCh)

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

func (s Stream) Close() error {
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

		s.out = NewLogWriter("", logx.New(texth.New(io.MultiWriter(f, os.Stderr))))
	case config.ToPrometheus:
		port, err := strconv.Atoi(cfg.OutputPath)
		if err != nil {
			port = 0
		}

		prom, err := NewPromWriter(port)
		if err != nil {
			return nil, err
		}

		s.out = prom
	default:
		s.out = NewLogWriter("", logx.New(texth.New(os.Stderr)))
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
	}

	return s, nil
}
