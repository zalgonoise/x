package stream

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusPeak struct {
	metric prometheus.Gauge
}

func (p PrometheusPeak) Write(v []int) (n int, err error) {
	for i := range v {
		p.metric.Set(float64(v[i]))
	}
	return len(v), nil
}

func (p PrometheusPeak) WriteItem(v int) error {
	p.metric.Set(float64(v))
	return nil
}

func NewPromPeak() PrometheusPeak {
	p := PrometheusPeak{promauto.NewGauge(prometheus.GaugeOpts{
		Name: "peak_value",
		Help: "input signal's peak value",
	})}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":13088", nil)
	}()

	return p
}

type PrometheusThreshold struct {
	metric prometheus.Gauge
}

func (p PrometheusThreshold) Write(v []int) (n int, err error) {
	for range v {
		p.metric.Inc()
	}
	return len(v), nil
}

func (p PrometheusThreshold) WriteItem(v int) error {
	p.metric.Inc()
	return nil
}

func NewPromThreshold() PrometheusThreshold {
	p := PrometheusThreshold{promauto.NewGauge(prometheus.GaugeOpts{
		Name: "over_threshold",
		Help: "input signal's peak value is over threshold",
	})}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":13089", nil)
	}()

	return p
}
