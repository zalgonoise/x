package stream

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zalgonoise/gio"
)

// DynamicThreshold is a custom metrics type that allows adding Prometheus collectors dynamically, in this case for
// adding peak threshold detectors to a monitor-mode setup
type DynamicThreshold struct {
	Peak      int
	threshold prometheus.Counter
}

// NewDynamicThreshold creates a DynamicThreshold with index `idx`, that increments a counter when the peak value is
// over `peak`
func NewDynamicThreshold(idx, peak int) (gio.ItemWriter[int], prometheus.Collector) {
	dt := DynamicThreshold{
		peak,
		prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: fmt.Sprintf("over_threshold_%d", idx),
				Help: fmt.Sprintf("amount of times the signal's peak level is over %d", peak),
			},
		),
	}

	return dt, dt.threshold
}

// Write implements the gio.Writer interface
//
// Its purpose is to expose a general means of writing incoming peak level values
// to a destination; in this case a prometheus.Gauge
func (p DynamicThreshold) Write(v []int) (n int, err error) {
	for i := range v {
		if err := p.WriteItem(v[i]); err != nil {
			return i, err
		}
	}
	return len(v), nil
}

// WriteItem implements the gio.ItemWriter interface
//
// Its purpose is to expose a general means of writing incoming peak level values
// to a destination; in this case a prometheus.Gauge
func (p DynamicThreshold) WriteItem(v int) error {
	if (p.Peak > 0 && v > p.Peak) || (p.Peak < 0 && v < p.Peak) {
		p.threshold.Inc()
	}
	return nil
}

// PrometheusPeak is an int Writer for registering PCM peak level items on Monitor Mode, into a prometheus.Gauge
type PrometheusPeak struct {
	peakValues prometheus.Gauge
	thresholds []gio.ItemWriter[int]
}

// Write implements the gio.Writer interface
//
// Its purpose is to expose a general means of writing incoming peak level values
// to a destination; in this case a prometheus.Gauge
func (p PrometheusPeak) Write(v []int) (n int, err error) {
	for i := range v {
		p.peakValues.Set(float64(v[i]))
	}
	return len(v), nil
}

// WriteItem implements the gio.ItemWriter interface
//
// Its purpose is to expose a general means of writing incoming peak level values
// to a destination; in this case a prometheus.Gauge
func (p PrometheusPeak) WriteItem(v int) error {
	p.peakValues.Set(float64(v))
	return nil
}

// NewPromPeak creates a PrometheusPeak
func NewPromPeak(port int, peaks ...int) (PrometheusPeak, []gio.ItemWriter[int], error) {
	var thresholdWriters []gio.ItemWriter[int]
	var thresholdCollectors []prometheus.Collector
	for i := range peaks {
		w, c := NewDynamicThreshold(i, peaks[i])
		thresholdWriters = append(thresholdWriters, w)
		thresholdCollectors = append(thresholdCollectors, c)
	}

	p := PrometheusPeak{
		peakValues: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "peak_value",
				Help: "input signal's peak value",
			},
		),
		thresholds: thresholdWriters,
	}
	err := prometheus.DefaultRegisterer.Register(p.peakValues)
	if err != nil {
		return p, thresholdWriters, err
	}
	for i := range p.thresholds {
		err := prometheus.DefaultRegisterer.Register(thresholdCollectors[i])
		if err != nil {
			return p, thresholdWriters, err
		}
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(
			fmt.Sprintf(":%d", port), nil,
		); err != nil {
			// probably the port is taken, and it's best to interrupt runtime
			panic(err)
		}
	}()

	return p, thresholdWriters, nil
}

// PrometheusThreshold is an int Writer for registering PCM peak level items on Filter Mode, when it surpasses
// the set peak, into a prometheus.Gauge
type PrometheusThreshold struct {
	metric prometheus.Gauge
}

// Write implements the gio.Writer interface
//
// Its purpose is to expose a general means of writing incoming peak level values
// to a destination; in this case a prometheus.Gauge
func (p PrometheusThreshold) Write(v []int) (n int, err error) {
	for range v {
		p.metric.Inc()
	}
	return len(v), nil
}

// WriteItem implements the gio.ItemWriter interface
//
// Its purpose is to expose a general means of writing incoming peak level values
// to a destination; in this case a prometheus.Gauge
func (p PrometheusThreshold) WriteItem(_ int) error {
	p.metric.Inc()
	return nil
}

// NewPromThreshold creates a PrometheusThreshold
func NewPromThreshold(port int) PrometheusThreshold {
	p := PrometheusThreshold{
		promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "over_threshold",
				Help: "input signal's peak value is over threshold",
			},
		),
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(
			fmt.Sprintf(":%d", port), nil,
		); err != nil {
			// panic since probably the port is taken, and it's best to interrupt runtime
			panic(err)
		}
	}()

	return p
}
