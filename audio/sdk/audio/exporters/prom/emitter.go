package prom

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zalgonoise/x/audio/fft"
	"github.com/zalgonoise/x/audio/sdk/audio"
	"github.com/zalgonoise/x/audio/sdk/audio/exporters"
	"github.com/zalgonoise/x/audio/wav/header"
	"github.com/zalgonoise/x/cfg"
)

const (
	defaultTimeout = 15 * time.Second
	defaultPort    = 13080

	expectedMaxFreqString = "22000"
	expectedMaxFreqLen    = len(expectedMaxFreqString)
)

type emitter struct {
	peaks   prometheus.Gauge
	spectra *prometheus.HistogramVec

	server *http.Server
}

func (e emitter) EmitPeaks(value float64) {
	e.peaks.Set(value)
}

func (e emitter) EmitSpectrum(values []fft.FrequencyPower) {
	for i := range values {
		e.spectra.
			WithLabelValues(minLen(strconv.Itoa(values[i].Freq), expectedMaxFreqLen)).
			Observe(values[i].Mag)
	}
}

func (e emitter) Shutdown(ctx context.Context) error {
	return e.server.Shutdown(ctx)
}

func ToProm(port int, options ...cfg.Option[exporters.Config]) (audio.Exporter, error) {
	e := emitter{
		peaks: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "audio",
			Name:      "peak_value",
			Help:      "input signal's peak value",
		}),
		spectra: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "audio",
			Name:      "spectrum_value",
			Help:      "input signal's frequency value",
		}, []string{"frequency"}),
	}

	reg, err := newRegistry(e)
	if err != nil {
		return audio.NoOpExporter[*header.Header](), err
	}

	e.server = newServer(port, reg)

	return exporters.NewExporter(e, options...)
}

func newRegistry(exporter emitter) (*prometheus.Registry, error) {
	reg := prometheus.NewRegistry()

	for _, metric := range []prometheus.Collector{
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{
			ReportErrors: false,
		}),
		exporter.peaks,
		exporter.spectra,
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
