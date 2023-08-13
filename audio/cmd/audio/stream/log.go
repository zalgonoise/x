package stream

import (
	"context"
	"log/slog"
	"time"

	"github.com/zalgonoise/x/audio/fft"
)

const (
	defaultMessage    = "new value registered"
	defaultTickerFreq = 200 * time.Millisecond
)

type LogWriter struct {
	msg string

	peakReg    *MaxRegistry[float64]
	freqReg    *MaxRegistry[fft.FrequencyPower]
	freqBucket *bucket[int]

	logger *slog.Logger
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
	w.logger.Info(w.msg, slog.Float64("power", data))
}

func (w LogWriter) setPeakFreq(frequency int, magnitude float64) {
	w.logger.Info(w.msg,
		slog.String("freq_bucket", w.freqBucket.Get(frequency)),
		slog.Int("frequency", frequency),
		slog.Float64("magnitude", magnitude),
	)
}

func (w LogWriter) flush() {
	if peak := w.peakReg.Flush(); peak > 0.0 {
		w.setPeakValue(peak)
	}
	if freq := w.freqReg.Flush(); freq.Freq > 0 {
		w.setPeakFreq(freq.Freq, freq.Mag)
	}
}

func NewLogWriter(message string, logger *slog.Logger) LogWriter {
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

		freqBucket: newBucket(frequencyValues, frequencyLabels, func(i, j int) bool { return i < j }),
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
