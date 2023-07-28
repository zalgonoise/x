package stream

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/zalgonoise/x/audio/fft"
)

const defaultTimeout = 5 * time.Second

type PromWriter struct {
	*Metrics
	*http.Server

	done context.CancelFunc
}

func (w PromWriter) Close() error {
	defer w.done()
	ctx, done := context.WithTimeout(context.Background(), defaultTimeout)
	defer done()

	return w.Server.Shutdown(ctx)
}

type Metrics struct {
	peakValues     prometheus.Gauge
	spectrumValues *prometheus.GaugeVec

	peakReg     *MaxRegistry[float64]
	spectrumReg BucketRegistry[fft.FrequencyPower, map[string]fft.FrequencyPower]
}

func (m Metrics) SetPeakValue(data float64) error {
	m.peakReg.Register(data)

	return nil
}

func (m Metrics) SetPeakFreq(frequency int, magnitude float64) (err error) {
	m.spectrumReg.Register(fft.FrequencyPower{Freq: frequency, Mag: magnitude})

	return nil
}

func (m Metrics) registry() (*prometheus.Registry, error) {
	reg := prometheus.NewRegistry()

	for _, metric := range []prometheus.Collector{
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{
			ReportErrors: false,
		}),
		m.peakValues,
		m.spectrumValues,
	} {
		if err := reg.Register(metric); err != nil {
			return nil, err
		}
	}

	return reg, nil
}

func (m Metrics) setPeakFreq(frequencyLabel string, magnitude float64) {
	m.spectrumValues.WithLabelValues(frequencyLabel).Set(magnitude)
}

func (m Metrics) setPeakValue(data float64) {
	m.peakValues.Set(data)
}

func (m Metrics) flush() {
	if peak := m.peakReg.Flush(); peak > 0.0 {
		m.setPeakValue(peak)
	}

	if spectrum := m.spectrumReg.Flush(); len(spectrum) > 0 {
		for k, v := range spectrum {
			m.setPeakFreq(k, v.Mag)
		}
	}
}

func NewMetrics() (*Metrics, error) {
	spectrumReg, err := NewBucketRegistry[fft.FrequencyPower, map[string]fft.FrequencyPower](
		defaultLabels[fft.FrequencyPower](
			func(i int) fft.FrequencyPower { return fft.FrequencyPower{Freq: i} },
			func(power fft.FrequencyPower) int { return power.Freq },
		),
	)
	if err != nil {
		return nil, err
	}

	return &Metrics{
		peakValues: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "audio",
			Name:      "peak_value",
			Help:      "input signal's peak value",
		}),
		spectrumValues: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "audio",
			Name:      "spectrum_value",
			Help:      "input signal's peak frequency value",
		}, []string{"frequency"}),

		peakReg: NewMaxRegistry(func(i, j float64) bool {
			return i < j
		}),
		spectrumReg: spectrumReg,
	}, nil
}

func NewServer(addr string, registry *prometheus.Registry) *http.Server {
	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		Registry: registry,
	}))

	server := &http.Server{
		Handler:      mux,
		Addr:         addr,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	return server
}

func NewPromWriter(addr string) (*PromWriter, error) {
	ctx, cancel := context.WithCancel(context.Background())

	m, err := NewMetrics()
	if err != nil {
		cancel()

		return nil, err
	}

	w := &PromWriter{
		Metrics: m,
		done:    cancel,
	}

	go func(ctx context.Context) {
		ticker := time.NewTicker(defaultTickerFreq)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				w.flush()

				return
			case <-ticker.C:
				w.flush()
			}
		}
	}(ctx)

	reg, err := w.Metrics.registry()
	if err != nil {
		cancel()

		return nil, err
	}

	w.Server = NewServer(addr, reg)

	return w, nil
}
