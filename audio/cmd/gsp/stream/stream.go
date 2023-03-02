package stream

import (
	"io"
	"os"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/logx"
	"github.com/zalgonoise/logx/handlers/texth"
	"github.com/zalgonoise/logx/level"
	"github.com/zalgonoise/x/audio/cmd/gsp/conf"
	"github.com/zalgonoise/x/audio/cmd/gsp/mode"
	"github.com/zalgonoise/x/audio/wav"
)

func New(cfg *conf.Config, r io.Reader) (*wav.WavBuffer, error) {
	w := wav.NewStream(r)

	switch cfg.Mode {
	case mode.Monitor:
		monitorMode(cfg, w)
	case mode.Filter:
		err := filterMode(cfg, w)
		if err != nil {
			return nil, err
		}
	case mode.Record:
		err := recordMode(cfg, w)
		if err != nil {
			return nil, err
		}
	}
	return w, nil
}

func monitorMode(cfg *conf.Config, w *wav.WavBuffer) {
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

func recordMode(cfg *conf.Config, w *wav.WavBuffer) error {
	output, err := os.Create(*cfg.Dir)
	if err != nil {
		return err
	}
	w.WithFilter(
		wav.FlushFor(output, *cfg.Dur),
	)
	return nil
}

func filterMode(cfg *conf.Config, w *wav.WavBuffer) error {
	output, err := os.Create(*cfg.Dir)
	if err != nil {
		return err
	}
	w.WithFilter(
		wav.LevelThreshold(*cfg.Peak, wav.FlushFor(output, *cfg.Dur)),
	)
	return nil
}
