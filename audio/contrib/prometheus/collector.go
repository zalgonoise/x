// Package prometheus exposes the measurements of an sdk/metric.ManualReader to Prometheus: every scrape collects the
// reader, and every series is reported as a gauge named after its instrument, with a "channel" label (CLAUDE.md §7 Q9).
// Spectra are reported as one gauge per channel and frequency, with a "frequency" label (the band's nominal centre, in
// Hz): give spectral instruments bands (metric.WithBands), as every FFT bin would be a series of its own.
//
// It lives in contrib/ because it is the only package of the module that imports the Prometheus client (§5.4).
package prometheus

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/zalgonoise/x/audio/sdk/metric"
)

// Labels: the channel of every series, and the frequency of spectra; mixdownLabel is the channel of a Measurement taken
// on the mixdown.
const (
	channelLabelName   = "channel"
	frequencyLabelName = "frequency"
	mixdownLabel       = "mix"
)

// ErrNameConflict is reported to the scrape for an instrument whose metric name another instrument already uses (once
// sanitised: "audio.peak" and "audio_peak" are both audio_peak). The first instrument keeps the name.
var ErrNameConflict = errors.New("prometheus: instruments share a metric name")

// Collector is a prometheus.Collector over a ManualReader. Collection is delta, so the Collector remembers the last
// value of each series and keeps reporting it until a newer one arrives: a scrape with no new audio repeats the last
// known levels instead of dropping them. When an instrument does report, its new series replace all of its previous
// ones: a channel, band or bin it no longer reports (a mono stream after a stereo one, a new sample rate) disappears.
type Collector struct {
	reader    *metric.ManualReader
	namespace string

	mu    sync.Mutex // also serialises scrapes, so that deltas apply in order
	last  map[seriesKey]float64
	keys  []seriesKey
	help  map[descKey]*prometheus.Desc
	names map[string]string // metric name → the instrument that owns it
}

// descKey identifies a metric family's descriptor: scalars and spectra have different labels.
type descKey struct {
	name     string
	spectrum bool
}

type seriesKey struct {
	name      string
	channel   string
	frequency string // spectra only
}

// Option configures a Collector.
type Option func(*Collector)

// WithNamespace prefixes every metric name (e.g. "radio" gives radio_audio_peak).
func WithNamespace(ns string) Option { return func(c *Collector) { c.namespace = ns } }

// NewCollector returns a Collector over reader. Register it with a prometheus.Registerer.
func NewCollector(reader *metric.ManualReader, opts ...Option) *Collector {
	c := &Collector{
		reader: reader,
		last:   map[seriesKey]float64{},
		help:   map[descKey]*prometheus.Desc{},
		names:  map[string]string{},
	}

	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}

	return c
}

// Describe implements prometheus.Collector. The Collector is unchecked: series appear as instruments produce them.
func (c *Collector) Describe(chan<- *prometheus.Desc) {}

// Collect implements prometheus.Collector.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.mu.Lock()
	defer c.mu.Unlock()

	batch, err := c.reader.Collect(context.Background())
	if err != nil {
		ch <- prometheus.NewInvalidMetric(prometheus.NewDesc("audio_collect_error", "Collecting audio metrics failed", nil,
			nil), err)
	}

	// the latest measurement of each instrument and channel: a later window (a restart in the same delta) replaces it
	fresh := map[string]map[string][]seriesValue{}
	conflicts := map[string]bool{}

	for _, m := range batch {
		name, channel := c.metricName(m.Name), channelLabel(m.Channel)

		if owner, ok := c.names[name]; ok && owner != m.Name {
			if !conflicts[m.Name] {
				conflicts[m.Name] = true
				ch <- prometheus.NewInvalidMetric(prometheus.NewDesc(name, "Conflicting audio measurement", nil, nil),
					fmt.Errorf("%w: %q and %q are both %s", ErrNameConflict, owner, m.Name, name))
			}

			continue
		}

		c.names[name] = m.Name

		if fresh[name] == nil {
			fresh[name] = map[string][]seriesValue{}
		}

		var values []seriesValue

		if m.Kind == metric.KindSpectrum {
			for _, b := range m.Spectrum {
				values = append(values, seriesValue{
					key:   seriesKey{name: name, channel: channel, frequency: frequencyLabel(b.Frequency)},
					value: b.Value,
				})
			}
		} else {
			values = append(values, seriesValue{key: seriesKey{name: name, channel: channel}, value: m.Value})
		}

		fresh[name][channel] = values
	}

	c.replace(fresh)

	for _, k := range c.keys {
		labels := []string{k.channel}
		if k.frequency != "" {
			labels = append(labels, k.frequency)
		}

		// a scalar and a spectrum whose names sanitise alike fail the scrape (inconsistent labels) instead of panicking
		m, err := prometheus.NewConstMetric(c.desc(k), prometheus.GaugeValue, c.last[k], labels...)
		if err != nil {
			m = prometheus.NewInvalidMetric(c.desc(k), err)
		}

		ch <- m
	}
}

type seriesValue struct {
	key   seriesKey
	value float64
}

// replace sets the series of every instrument in fresh, by metric name and channel, and drops the instrument's other
// series.
func (c *Collector) replace(fresh map[string]map[string][]seriesValue) {
	if len(fresh) == 0 {
		return
	}

	current := map[seriesKey]bool{}

	for _, channels := range fresh {
		for _, values := range channels {
			for _, v := range values {
				current[v.key] = true
			}
		}
	}

	c.keys = slices.DeleteFunc(c.keys, func(k seriesKey) bool {
		if _, reported := fresh[k.name]; !reported || current[k] {
			return false
		}

		delete(c.last, k)

		return true
	})

	for _, channels := range fresh {
		for _, values := range channels {
			for _, v := range values {
				if _, ok := c.last[v.key]; !ok {
					c.keys = append(c.keys, v.key)
				}

				c.last[v.key] = v.value
			}
		}
	}
}

// desc returns the descriptor of k's metric family: spectra have a frequency label as well as the channel.
func (c *Collector) desc(k seriesKey) *prometheus.Desc {
	dk := descKey{name: k.name, spectrum: k.frequency != ""}
	if desc, ok := c.help[dk]; ok {
		return desc
	}

	desc := prometheus.NewDesc(k.name, "Audio measurement "+k.name+", per channel", []string{channelLabelName}, nil)
	if dk.spectrum {
		desc = prometheus.NewDesc(k.name, "Audio spectrum "+k.name+" (linear amplitude, 1 is a full-scale sine), per "+
			"channel and frequency in Hz", []string{channelLabelName, frequencyLabelName}, nil)
	}

	c.help[dk] = desc

	return desc
}

func (c *Collector) metricName(name string) string {
	if c.namespace != "" {
		name = c.namespace + "_" + name
	}

	return sanitize(name)
}

// sanitize maps an instrument name to a valid Prometheus metric name: [a-zA-Z_:][a-zA-Z0-9_:]*.
func sanitize(name string) string {
	var b strings.Builder

	for i, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r == '_', r == ':':
			b.WriteRune(r)
		case r >= '0' && r <= '9' && i > 0:
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}

	return b.String()
}

// frequencyLabel formats a frequency in Hz with as few digits as needed: 31.5, 1000, 12500.
func frequencyLabel(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func channelLabel(ch int) string {
	if ch == metric.Mixdown {
		return mixdownLabel
	}

	return strconv.Itoa(ch)
}
