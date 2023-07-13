package stream

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PromWriter struct {
	*Metrics
	MetricsServer
}

type Metrics struct {
	peakValues prometheus.Gauge
}

func (m Metrics) SetPeakValue(data float64) error {
	m.peakValues.Set(data)

	return nil
}

func (m Metrics) SetPeakValues(data []float64) error {
	for i := range data {
		if err := m.SetPeakValue(data[i]); err != nil {
			return err
		}
	}

	return nil
}

func (m Metrics) Registry() (*prometheus.Registry, error) {
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

func NewMetrics() *Metrics {
	return &Metrics{
		peakValues: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "peak_value",
			Help: "input signal's peak value",
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

func NewServer(port int, registry *prometheus.Registry) MetricsServer {
	if port < 0 {
		port = 0
	}

	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		Registry: registry,
	}))

	server := &http.Server{
		Handler:      mux,
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	return MetricsServer{server: server}
}

func NewPromWriter(port int) (*PromWriter, error) {
	w := &PromWriter{
		Metrics: NewMetrics(),
	}

	reg, err := w.Metrics.Registry()
	if err != nil {
		return nil, err
	}

	w.MetricsServer = NewServer(port, reg)

	return w, w.MetricsServer.ListenAndServe()
}
