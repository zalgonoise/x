package stream

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/zalgonoise/gio"

	"github.com/zalgonoise/x/audio/wav"
	"github.com/zalgonoise/x/audio/wav/fft"
)

// New creates a WAV stream from the input Config `cfg` and io.Reader `r`
//
// It returns a pointer to a wav.Buffer and an error if raised
func New(cfg *Config, r io.Reader) (*wav.WavBuffer, error) {
	w := wav.NewStream(r).
		Ratio(cfg.BufferSize).
		BlockSize(cfg.BlockSize)

	switch cfg.Mode {
	case Monitor:
		if err := monitorMode(cfg, w); err != nil {
			return nil, err
		}
	case Filter:
		if err := filterMode(cfg, w); err != nil {
			return nil, err
		}
	case Record:
		if err := recordMode(cfg, w); err != nil {
			return nil, err
		}
	case Analyze:
		if err := analyzerMode(cfg, w); err != nil {
			return nil, err
		}
	}
	return w, nil
}

func monitorWriter(cfg *Config) (gio.ItemWriter[int], []gio.ItemWriter[int], error) {
	if cfg.Prom {
		w, peaksW, err := NewPromPeak(cfg.Port, cfg.Peak...)
		if err != nil {
			return nil, nil, err
		}
		return w, peaksW, nil
	}
	return NewLoggerPeak(cfg.Peak...), nil, nil
}

func monitorMode(cfg *Config, w *wav.WavBuffer) error {
	writer, peaksWriters, err := monitorWriter(cfg)
	if err != nil {
		return err
	}
	var maxCh = make(chan int)

	go func() {
		for {
			select {
			case value := <-maxCh:
				_ = writer.WriteItem(value)
				for idx := range peaksWriters {
					_ = peaksWriters[idx].WriteItem(value)
				}
			}
		}
	}()

	w.WithFilter(
		wav.MaxValues(maxCh),
	)
	return nil
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
		return NewPromThreshold(cfg.Port)
	}
	return NewLoggerThreshold(cfg.Peak[0])
}

func filterMode(cfg *Config, w *wav.WavBuffer) error {
	if len(cfg.Peak) == 0 {
		return ErrEmptyThreshold
	}
	writer := filterWriter(cfg)
	w.WithFilter(
		wav.LevelThresholdFn(
			cfg.Peak[0],
			func(v int) { _ = writer.WriteItem(v) },
			wav.FlushToFileFor(*cfg.Dir, *cfg.RecTime),
		),
	)
	return nil
}

func analyzerMode(cfg *Config, w *wav.WavBuffer) error {
	var spectrumCh = make(chan []fft.FrequencyPower)

	var bs = fft.Block128
	// work with a BlockSize half the size of the ring filter's, if configured
	if cfg.BlockSize >= int(fft.Block16) {
		bs = fft.AsBlock(cfg.BlockSize / 2)
	}

	w.WithFilter(
		wav.Spectrum(bs, spectrumCh),
	)

	go func() {
		err := NewEQ(spectrumCh)
		if err != nil {
			panic(err)
		}
	}()

	return nil
}
