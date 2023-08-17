package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/zalgonoise/x/audio/cmd/audio/config"
	"github.com/zalgonoise/x/audio/cmd/audio/stream"
)

func main() {
	err, code := runV2()
	if err != nil {
		slog.Error(
			"audio/gsp: runtime error",
			slog.String("error", err.Error()),
		)
	}

	os.Exit(code)
}

func run() (error, int) {
	cfg, err := config.WithDefaults()
	if err != nil {
		return err, 1
	}

	s, err := stream.New(cfg)
	if err != nil {
		return err, 1
	}

	ctx := context.Background()

	err = s.Run(ctx)
	if err != nil {
		return err, 1
	}

	err = s.Close()
	if err != nil {
		return err, 1
	}

	return nil, 0
}

func runV2() (error, int) {
	cfg, err := config.WithDefaults()
	if err != nil {
		return err, 1
	}

	consumer, err := stream.NewHTTPConsumer(cfg.URL, cfg.Duration)
	if err != nil {
		return err, 1
	}

	exporter := stream.NewLogExporter(os.Stderr)

	processor, err := stream.NewPCMProcessor(
		exporter,
		&stream.ProcessorConfig{Size: 64},
		stream.ProcessPeaks, stream.ProcessSpectrum,
	)
	if err != nil {
		return err, 1
	}

	ctx := context.Background()
	reader, err := consumer.Consume(ctx)
	if err != nil {
		return err, 1
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	go processor.Process(ctx, reader)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err(), 0
		case err := <-processor.Err():
			return err, 1
		}
	}
}
