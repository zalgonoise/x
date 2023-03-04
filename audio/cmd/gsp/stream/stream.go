package stream

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/logx"
	"github.com/zalgonoise/logx/handlers/texth"
	"github.com/zalgonoise/logx/level"
	"github.com/zalgonoise/x/audio/wav"
)

func New(cfg *Config, r io.Reader) (*wav.WavBuffer, error) {
	w := wav.NewStream(r)
	w.Ratio(cfg.BufferSize)

	switch cfg.Mode {
	case Monitor:
		monitorMode(cfg, w)
	case Filter:
		err := filterMode(cfg, w)
		if err != nil {
			return nil, err
		}
	case Record:
		err := recordMode(cfg, w)
		if err != nil {
			return nil, err
		}
	}
	return w, nil
}

func monitorMode(cfg *Config, w *wav.WavBuffer) {
	logger := logx.New(texth.New(os.Stdout))
	var maxCh = make(chan int)
	go func() {
		for i := range maxCh {
			logger.Log(level.Info, "peak level", attr.Int("value", i))
		}
	}()

	w.WithFilter(
		wav.MaxValues(maxCh),
	)
}

func recordMode(cfg *Config, w *wav.WavBuffer) error {

	output, err := os.Create(fmt.Sprintf("%s_%s.wav", *cfg.Dir, time.Now().Format(time.RFC3339)))
	if err != nil {
		return err
	}
	w.WithFilter(
		wav.FlushFor(output, *cfg.RecTime),
	)
	return nil
}

func filterMode(cfg *Config, w *wav.WavBuffer) error {
	w.WithFilter(
		wav.LevelThreshold(*cfg.Peak, wav.FlushToFileFor(*cfg.Dir, *cfg.RecTime)),
	)
	return nil
}
