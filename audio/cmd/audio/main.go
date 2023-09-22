package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/zalgonoise/x/audio/fft"
	"github.com/zalgonoise/x/audio/sdk/audio"
	"github.com/zalgonoise/x/audio/sdk/audio/compactors"
	"github.com/zalgonoise/x/audio/sdk/audio/consumers/httpaudio"
	"github.com/zalgonoise/x/audio/sdk/audio/exporters/stdout"
	"github.com/zalgonoise/x/audio/sdk/audio/processors"
	"github.com/zalgonoise/x/audio/sdk/audio/registries/batchreg"
)

func main() {
	err, code := run()
	if err != nil {
		slog.Error(
			"audio: runtime error",
			slog.String("error", err.Error()),
		)
	}

	os.Exit(code)
}

func run() (error, int) {
	consumer, err := httpaudio.New(
		httpaudio.WithTarget("http://192.168.10.12:8080/audio.wav"),
		httpaudio.WithTimeout(3*time.Minute),
	)

	if err != nil {
		return err, 1
	}

	exporter, err := stdout.ToLogger(
		stdout.WithPeaks(),
		stdout.WithBatchedPeaks(
			batchreg.WithBatchSize[float64](256),
			batchreg.WithFlushFrequency[float64](500*time.Millisecond),
			batchreg.WithCompactor[float64](compactors.Max[float64]),
		),
		stdout.WithSpectrum(128),
		stdout.WithBatchedSpectrum(
			batchreg.WithBatchSize[[]fft.FrequencyPower](256),
			batchreg.WithFlushFrequency[[]fft.FrequencyPower](500*time.Millisecond),
			batchreg.WithCompactor[[]fft.FrequencyPower](compactors.MaxSpectra),
		),
	)

	if err != nil {
		return err, 1
	}

	proc := processors.NewPCM(exporter)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	r, err := consumer.Consume(ctx)
	if err != nil {
		return err, 1
	}

	go proc.Process(ctx, r)

	errs := proc.Err()

	defer audio.Shutdown(ctx, 15*time.Second, consumer, proc)

	for {
		select {
		case <-ctx.Done():
			return nil, 0
		case err, ok := <-errs:
			if !ok || err == nil || errors.Is(err, processors.ErrHaltSignal) {
				return nil, 0
			}

			return err, 1
		}
	}
}
