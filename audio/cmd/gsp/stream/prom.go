package stream

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusPeak is an int Writer for registering PCM peak level items on Monitor Mode, into a prometheus.Gauge
type PrometheusPeak struct {
	metric prometheus.Gauge
}

// Write implements the gio.Writer interface
//
// Its purpose is to expose a general means of writing incoming peak level values
// to a destination; in this case a prometheus.Gauge
func (p PrometheusPeak) Write(v []int) (n int, err error) {
	for i := range v {
		p.metric.Set(float64(v[i]))
	}
	return len(v), nil
}

// WriteItem implements the gio.ItemWriter interface
//
// Its purpose is to expose a general means of writing incoming peak level values
// to a destination; in this case a prometheus.Gauge
func (p PrometheusPeak) WriteItem(v int) error {
	p.metric.Set(float64(v))
	return nil
}

// NewPromPeak creates a PrometheusPeak
func NewPromPeak() PrometheusPeak {
	p := PrometheusPeak{promauto.NewGauge(prometheus.GaugeOpts{
		Name: "peak_value",
		Help: "input signal's peak value",
	})}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":13088", nil); err != nil {
			// panic since probably the port is taken, and it's best to interrupt runtime
			panic(err)
		}
	}()

	return p
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
func NewPromThreshold() PrometheusThreshold {
	p := PrometheusThreshold{promauto.NewGauge(prometheus.GaugeOpts{
		Name: "over_threshold",
		Help: "input signal's peak value is over threshold",
	})}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":13089", nil); err != nil {
			// panic since probably the port is taken, and it's best to interrupt runtime
			panic(err)
		}
	}()

	return p
}
