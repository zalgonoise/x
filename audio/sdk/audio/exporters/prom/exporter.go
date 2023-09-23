package prom

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zalgonoise/x/audio/fft"
	"github.com/zalgonoise/x/audio/sdk/audio"
	"github.com/zalgonoise/x/audio/sdk/audio/compactors"
	"github.com/zalgonoise/x/audio/sdk/audio/extractors"
	"github.com/zalgonoise/x/audio/sdk/audio/registries/batchreg"
	"github.com/zalgonoise/x/audio/sdk/audio/registries/unitreg"
	"github.com/zalgonoise/x/audio/wav/header"
	"github.com/zalgonoise/x/cfg"
)

const (
	defaultTimeout = 15 * time.Second
	defaultPort    = 13080

	expectedMaxFreqString = "22000"
	expectedMaxFreqLen    = len(expectedMaxFreqString)
)

type promExporter struct {
	config Config

	peaks    audio.Collector[float64]
	spectrum audio.Collector[[]fft.FrequencyPower]

	peakValues     prometheus.Gauge
	spectrumValues *prometheus.HistogramVec

	server *http.Server

	logger *slog.Logger

	cancel context.CancelFunc
}

func (e promExporter) Export(h *header.Header, data []float64) error {
	return errors.Join(
		e.peaks.Collect(h, data),
		e.spectrum.Collect(h, data),
	)
}

func (e promExporter) ForceFlush() error {
	return errors.Join(
		e.peaks.ForceFlush(),
		e.spectrum.ForceFlush(),
	)
}

func (e promExporter) Shutdown(ctx context.Context) error {
	e.logger.InfoContext(ctx, "prometheus exporter shutting down")
	e.cancel()

	return errors.Join(
		e.peaks.Shutdown(ctx),
		e.spectrum.Shutdown(ctx),
	)
}

func (e promExporter) export(ctx context.Context) {
	peaksValues := e.peaks.Load()
	spectrumValues := e.spectrum.Load()

	for {
		select {
		case <-ctx.Done():
			e.logger.InfoContext(ctx, "stopping exporter's routine")

			return
		case v, ok := <-peaksValues:
			if !ok {
				return
			}

			e.peakValues.Set(v)
		case v, ok := <-spectrumValues:
			if !ok {
				return
			}

			for i := range v {
				e.spectrumValues.
					WithLabelValues(minLen(strconv.Itoa(v[i].Freq), expectedMaxFreqLen)).
					Observe(v[i].Mag)
			}
		}
	}
}

func ToPrometheus(port int, options ...cfg.Option[Config]) (audio.Exporter, error) {
	config := cfg.Set[Config](defaultConfig, options...)

	ctx, cancel := context.WithCancel(context.Background())

	exporter := promExporter{
		config:   config,
		peaks:    newPeaksCollector(config),
		spectrum: newSpectrumCollector(config),
		peakValues: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "audio",
			Name:      "peak_value",
			Help:      "input signal's peak value",
		}),
		spectrumValues: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "audio",
			Name:      "spectrum_value",
			Help:      "input signal's frequency value",
		}, []string{"frequency"}),
		logger: slog.New(config.logHandler),
		cancel: cancel,
	}

	reg, err := newRegistry(exporter)
	if err != nil {
		return audio.NoOpExporter(), err
	}

	exporter.logger.InfoContext(ctx, "starting metrics server", slog.Int("port", port))
	exporter.server = newServer(port, reg)

	go exporter.export(ctx)

	return exporter, nil
}

func newRegistry(exporter promExporter) (*prometheus.Registry, error) {
	reg := prometheus.NewRegistry()

	for _, metric := range []prometheus.Collector{
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{
			ReportErrors: false,
		}),
		exporter.peakValues,
		exporter.spectrumValues,
	} {
		if err := reg.Register(metric); err != nil {
			return nil, err
		}
	}

	return reg, nil
}

func newServer(port int, reg *prometheus.Registry) *http.Server {
	if port < 0 {
		port = defaultPort
	}

	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		Registry: reg,
	}))

	server := &http.Server{
		Handler:      mux,
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  defaultTimeout,
		WriteTimeout: defaultTimeout,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	return server
}

func minLen(s string, size int) string {
	out := make([]byte, size)
	j := size - 1

	for i := len(s) - 1; i >= 0; i-- {
		out[j] = s[i]
		j--
	}

	for ; j >= 0; j-- {
		out[j] = '0'
	}

	return string(out)
}

func newPeaksCollector(config Config) audio.Collector[float64] {
	if !config.withPeaks {
		return audio.NoOpCollector[float64]()
	}

	if !config.batchedPeaks {
		return audio.NewCollector[float64](
			extractors.MaxPeak(),
			unitreg.New[float64](0),
		)
	}

	return audio.NewCollector[float64](
		extractors.MaxPeak(),
		batchreg.New[float64](config.batchedPeaksOptions...),
	)
}

func newSpectrumCollector(config Config) audio.Collector[[]fft.FrequencyPower] {
	if !config.withSpectrum {
		return audio.NoOpCollector[[]fft.FrequencyPower]()
	}

	if !config.batchedPeaks {
		return audio.NewCollector[[]fft.FrequencyPower](
			extractors.Spectrum(config.spectrumBlockSize, compactors.UpperSpectra),
			unitreg.New[[]fft.FrequencyPower](0),
		)
	}

	return audio.NewCollector[[]fft.FrequencyPower](
		extractors.Spectrum(config.spectrumBlockSize, compactors.UpperSpectra),
		batchreg.New[[]fft.FrequencyPower](config.batchedSpectrumOptions...),
	)
}
