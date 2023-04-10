package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/logx"
	"github.com/zalgonoise/logx/handlers/texth"

	"github.com/zalgonoise/x/audio/cmd/gsp/client"
	"github.com/zalgonoise/x/audio/cmd/gsp/stream"
)

func main() {
	err, code := run()
	if err != nil {
		logx.New(texth.New(os.Stderr)).Error(
			"audio/gsp: runtime error",
			attr.String("error", err.Error()),
		)
	}
	os.Exit(code)
}

func run() (error, int) {
	cfg, err := stream.ParseFlags()
	if err != nil {
		return fmt.Errorf("initializing configuration: %w", err), 1
	}

	c, cancel, err := client.New(cfg.URL, cfg.Dur)
	if err != nil {
		return fmt.Errorf("setting up HTTP client: %w", err), 1
	}

	res, err := c.Do()
	if err != nil {
		return fmt.Errorf("issuing HTTP request: %w", err), 1
	}

	wav, err := stream.New(cfg, res.Body)
	if err != nil {
		return fmt.Errorf("setting up a wav.WavBuffer for this audio stream: %w", err), 1
	}

	// start processing the audio from the HTTP stream
	// create a local context cancel func to exit gracefully (on sigint, for example)
	errCh := make(chan error)
	ctx, mainCancel := context.WithCancel(c.Context())
	go wav.Stream(ctx, errCh)

	for {
		select {
		case err := <-errCh:
			mainCancel()
			if !errors.Is(err, context.Canceled) {
				return fmt.Errorf("audio stream: %w", err), 1
			}
		case <-ctx.Done():
			// context timed out, exit gracefully
			mainCancel()
			defer cancel()
			return res.Body.Close(), cfg.ExitCode
		}
	}
}
