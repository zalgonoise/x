package stream

import (
	"context"
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
	MetricsServer

	done context.CancelFunc
}

func (w PromWriter) Close() error {
	defer w.done()
	ctx, done := context.WithTimeout(context.Background(), defaultTimeout)
	defer done()

	return w.MetricsServer.Shutdown(ctx)
}

type Metrics struct {
	peakValues     prometheus.Gauge
	spectrumValues prometheus.Gauge

	peakReg *MaxRegistry[float64]
	freqReg *MaxRegistry[fft.FrequencyPower]
}

func (m Metrics) SetPeakValue(data float64) error {
	m.peakReg.Register(data)

	return nil
}

func (m Metrics) SetPeakFreq(frequency int, magnitude float64) (err error) {
	m.freqReg.Register(fft.FrequencyPower{Freq: frequency, Mag: magnitude})

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
	} {
		if err := reg.Register(metric); err != nil {
			return nil, err
		}
	}

	return reg, nil
}

func (m Metrics) setPeakFreq(frequency int) {
	m.spectrumValues.Set(float64(frequency))
}

func (m Metrics) setPeakValue(data float64) {
	m.peakValues.Set(data)
}

func (m Metrics) flush() {
	if peak := m.peakReg.Flush(); peak > 0.0 {
		m.setPeakValue(peak)
	}
	if freq := m.freqReg.Flush(); freq.Freq > 0 {
		m.setPeakFreq(freq.Freq)
	}
}

func NewMetrics() *Metrics {
	return &Metrics{
		peakValues: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "peak_value",
			Help: "input signal's peak value",
		}),
		spectrumValues: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "spectrum_value",
			Help: "input signal's peak frequency value",
		}),

		peakReg: NewMaxRegistry(func(i, j float64) bool {
			return i < j
		}),
		freqReg: NewMaxRegistry(func(i, j fft.FrequencyPower) bool {
			return i.Mag < j.Mag
		}),
	}
}

type MetricsServer struct {
	server *http.Server
	err    error
}

func (s MetricsServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s MetricsServer) ListenAndServe() error {
	if s.err != nil {
		return s.err
	}

	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			s.err = err
		}
	}()

	return nil
}

func NewServer(addr string, registry *prometheus.Registry) MetricsServer {
	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		Registry: registry,
	}))

	server := &http.Server{
		Handler:      mux,
		Addr:         addr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	return MetricsServer{server: server}
}

func NewPromWriter(addr string) (*PromWriter, error) {
	ctx, cancel := context.WithCancel(context.Background())

	w := &PromWriter{
		Metrics: NewMetrics(),
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
		return nil, err
	}

	w.MetricsServer = NewServer(addr, reg)

	return w, nil
}
