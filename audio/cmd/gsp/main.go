package main

import (
	"flag"
	"os"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/logx"
	"github.com/zalgonoise/logx/handlers/texth"
	"github.com/zalgonoise/x/audio/cmd/gsp/client"
	"github.com/zalgonoise/x/audio/cmd/gsp/stream"
)

func ParseFlags() (*stream.Config, error) {
	url := flag.String("url", "", "the URL for the WAV audio stream")
	dur := flag.String("dur", "", "duration until the operation times out")
	recTime := flag.String("rec", "", "duration of each recording")
	mode := flag.String("mode", "monitor", "operation mode (monitor, record, filter)")
	peak := flag.Int("peak", 0, "filter peak value to trigger recording the incoming signal")
	dir := flag.String("d", "./sound_capture.wav", "the path to the destination file")
	term := flag.Bool("t", true, "print monitor values to standard out")
	bufferSize := flag.Float64("s", 1.0, "buffer size as a ratio in seconds. 1.0 cycles the buffer once every second. 0.5 cycles twice per second. 2.0 cycles every two seconds; etc.")

	flag.Parse()

	return stream.NewConfig(*url, *mode, *bufferSize, dur, recTime, dir, peak, *term)
}

func main() {
	logger := logx.New(texth.New(os.Stderr))

	cfg, err := ParseFlags()
	if err != nil {
		logger.Error("failed to parse CLI flags", attr.String("error", err.Error()))
		os.Exit(1)
	}

	c, cancel, err := client.New(cfg.URL, cfg.Dur)
	if err != nil {
		logger.Error("failed to setup HTTP client", attr.String("error", err.Error()))
		os.Exit(1)
	}
	res, err := c.Do()
	if err != nil {
		logger.Error("HTTP request raised an error", attr.String("error", err.Error()))
		os.Exit(1)
	}

	wav, err := stream.New(cfg, res.Body)
	if err != nil {
		logger.Error("failed to prepare WAV buffer", attr.String("error", err.Error()))
		os.Exit(1)
	}

	errCh := make(chan error)
	ctx := c.Context()
	go wav.Stream(ctx, errCh)

	select {
	case err := <-errCh:
		logger.Error("error raised while streaming", attr.String("error", err.Error()))
		os.Exit(1)
	case <-ctx.Done():
		res.Body.Close()
		cancel()
	}
}
