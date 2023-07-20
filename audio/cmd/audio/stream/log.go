package stream

import (
	"context"
	"time"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/logx"

	"github.com/zalgonoise/x/audio/fft"
)

const (
	defaultMessage    = "new value registered"
	defaultTickerFreq = 200 * time.Millisecond
)

type LogWriter struct {
	msg string

	peakReg *MaxRegistry[float64]
	freqReg *MaxRegistry[fft.FrequencyPower]

	logger logx.Logger
	done   context.CancelFunc
}

func (w LogWriter) SetPeakValue(data float64) (err error) {
	w.peakReg.Register(data)

	return nil
}

func (w LogWriter) SetPeakFreq(frequency int, magnitude float64) (err error) {
	w.freqReg.Register(fft.FrequencyPower{Freq: frequency, Mag: magnitude})

	return nil
}

func (w LogWriter) Close() error {
	defer w.done()

	return nil
}

func (w LogWriter) setPeakValue(data float64) {
	w.logger.Info(w.msg, attr.Float("power", data))
}

func (w LogWriter) setPeakFreq(frequency int) {
	w.logger.Info(w.msg,
		attr.Int("frequency", frequency),
	)
}

func (w LogWriter) flush() {
	if peak := w.peakReg.Flush(); peak > 0.0 {
		w.setPeakValue(peak)
	}
	if freq := w.freqReg.Flush(); freq.Freq > 0 {
		w.setPeakFreq(freq.Freq)
	}
}

func NewLogWriter(message string, logger logx.Logger) LogWriter {
	if message == "" {
		message = defaultMessage
	}

	ctx, cancel := context.WithCancel(context.Background())

	w := LogWriter{
		msg:    message,
		logger: logger,
		done:   cancel,

		peakReg: NewMaxRegistry(func(i, j float64) bool {
			return i < j
		}),
		freqReg: NewMaxRegistry(func(i, j fft.FrequencyPower) bool {
			return i.Mag < j.Mag
		}),
	}

	go func(ctx context.Context) {
		ticker := time.NewTicker(defaultTickerFreq)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				w.flush()

				return
			case <-ticker.C:
				w.flush()
			}
		}
	}(ctx)

	return w
}
