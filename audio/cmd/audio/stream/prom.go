package stream

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/zalgonoise/x/audio/fft"
)

const (
	defaultTimeout        = 5 * time.Second
	expectedMaxFreqString = "22000"
	expectedMaxFreqLen    = len(expectedMaxFreqString)
)

type PromWriter struct {
	*Metrics
	*http.Server
}

func (w PromWriter) Close() error {
	ctx, done := context.WithTimeout(context.Background(), defaultTimeout)
	defer done()

	return w.Server.Shutdown(ctx)
}

type Metrics struct {
	peakValues     prometheus.Gauge
	spectrumValues *prometheus.HistogramVec
}

func (m Metrics) SetPeakValue(data float64) error {
	m.peakValues.Set(data)

	return nil
}

func (m Metrics) ObserveFrequencies(frequencies []fft.FrequencyPower) (err error) {
	for i := range frequencies {
		m.spectrumValues.
			WithLabelValues(minLen(strconv.Itoa(frequencies[i].Freq), expectedMaxFreqLen)).
			Observe(frequencies[i].Mag)
	}

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

func NewMetrics() (*Metrics, error) {
	return &Metrics{
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
	m, err := NewMetrics()
	if err != nil {
		return nil, err
	}

	reg, err := m.registry()
	if err != nil {
		return nil, err
	}

	return &PromWriter{
		Metrics: m,
		Server:  NewServer(addr, reg),
	}, nil
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
