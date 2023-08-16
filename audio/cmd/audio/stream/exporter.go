package stream

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zalgonoise/x/audio/fft"
)

type PromExporter struct {
	peakValues     prometheus.Gauge
	spectrumValues *prometheus.HistogramVec

	server *http.Server
	err    error
}

// SetPeakValue registers the float64 `data` value as an audio peak
func (e *PromExporter) SetPeakValue(data float64) (err error) {
	e.peakValues.Set(data)

	return nil
}

// ObserveFrequencies keeps track of changes in the registered frequencies
func (e *PromExporter) ObserveFrequencies(frequencies []fft.FrequencyPower) (err error) {
	for i := range frequencies {
		e.spectrumValues.
			WithLabelValues(minLen(strconv.Itoa(frequencies[i].Freq), expectedMaxFreqLen)).
			Observe(frequencies[i].Mag)
	}

	return nil
}

// Shutdown gracefully stops the Exporter
func (e *PromExporter) Shutdown(ctx context.Context) (err error) {
	if e.err != nil {
		return err
	}

	return e.server.Shutdown(ctx)
}

func NewPromExporter(address string) (*PromExporter, error) {
	// initialize metrics components
	exporter := &PromExporter{
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
	}

	// register metrics to expose
	registry := prometheus.NewRegistry()
	for _, metric := range []prometheus.Collector{
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{
			ReportErrors: false,
		}),
		exporter.peakValues,
		exporter.spectrumValues,
	} {
		if err := registry.Register(metric); err != nil {
			return nil, err
		}
	}

	// create and initialize the metrics http.Server
	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		Registry: registry,
	}))

	exporter.server = &http.Server{
		Handler:      mux,
		Addr:         address,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		if err := exporter.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			exporter.err = err
		}
	}()

	return exporter, nil
}

type LogExporter struct {
	writer io.Writer
	logger *slog.Logger
}

// SetPeakValue registers the float64 `data` value as an audio peak
func (e *LogExporter) SetPeakValue(data float64) (err error) {
	e.logger.Info("audio peak value registered", slog.Float64("power", data))

	return nil
}

// ObserveFrequencies keeps track of changes in the registered frequencies
func (e *LogExporter) ObserveFrequencies(frequencies []fft.FrequencyPower) (err error) {
	var maximum fft.FrequencyPower

	for i := range frequencies {
		if frequencies[i].Mag > maximum.Mag {
			maximum = frequencies[i]
		}
	}

	e.logger.Info("audio peak frequency registered",
		slog.Int("frequency", maximum.Freq),
		slog.Float64("magnitude", maximum.Mag),
	)

	return nil
}

// Shutdown gracefully stops the Exporter
func (e *LogExporter) Shutdown(ctx context.Context) (err error) {
	if closer, ok := (e.writer).(io.Closer); ok {
		return closer.Close()
	}

	if closer, ok := (e.writer).(interface {
		Shutdown(context.Context) error
	}); ok {
		return closer.Shutdown(ctx)
	}

	return nil
}

func NewLogExporter(writer io.Writer) *LogExporter {
	return &LogExporter{
		writer: writer,
		logger: slog.New(slog.NewTextHandler(writer, &slog.HandlerOptions{
			AddSource: true,
		})),
	}
}
