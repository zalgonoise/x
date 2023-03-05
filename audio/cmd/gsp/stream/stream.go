package stream

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/zalgonoise/gio"
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

func monitorWriter(cfg *Config) gio.ItemWriter[int] {
	if cfg.Prom {
		return NewPromPeak()
	}
	return NewLoggerPeak()
}

func monitorMode(cfg *Config, w *wav.WavBuffer) {
	writer := monitorWriter(cfg)
	var maxCh = make(chan int)
	go func() {
		for i := range maxCh {
			_ = writer.WriteItem(i)
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

func filterWriter(cfg *Config) gio.ItemWriter[int] {
	if cfg.Prom {
		return NewPromThreshold()
	}
	return NewLoggerThreshold(*cfg.Peak)
}

func filterMode(cfg *Config, w *wav.WavBuffer) error {
	writer := filterWriter(cfg)
	w.WithFilter(
		wav.LevelThresholdFn(
			*cfg.Peak,
			func(v int) { _ = writer.WriteItem(v) },
			wav.FlushToFileFor(*cfg.Dir, *cfg.RecTime)),
	)
	return nil
}
