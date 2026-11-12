// Package metric is the analysis plane of the SDK, shaped after the OpenTelemetry metrics SDK (CLAUDE.md §5.1).
//
// A Meter is an sdk.Exporter. For every Buffer it runs each Instrument once per channel (or once on the mixdown) and
// records the value into a per-channel Aggregation. A Reader collects the aggregations: a PeriodicReader emits them
// every interval to an Emitter (stdout, logs, OTLP), and a ManualReader collects them on demand (Prometheus scrapes).
// Collection is delta: every Measurement covers the frames recorded since the previous collection.
//
// Instruments are scalar (NewInstrument: one value per buffer, such as Levels) or spectral (NewSpectrum: an amplitude
// spectrum per frame of a short-time Fourier transform, carried across buffers, optionally reduced to frequency bands,
// such as Spectra).
//
// Wrap a Meter in an sdk.QueuedExporter to keep measuring off the pipeline's goroutine.
package metric

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"sync"

	"github.com/zalgonoise/x/audio"
	"github.com/zalgonoise/x/audio/dsp/feature"
	"github.com/zalgonoise/x/audio/sdk"
)

// Mixdown is the Channel of a Measurement taken on the average of all channels (see OnMixdown).
const Mixdown = -1

// ErrInvalidInstrument is returned for an Instrument without a name, function or aggregation.
var ErrInvalidInstrument = errors.New("metric: invalid instrument")

// Kind is the kind of value a Measurement holds.
type Kind uint8

const (
	// KindScalar is a single value, in Measurement.Value (instruments built with NewInstrument).
	KindScalar Kind = iota
	// KindSpectrum is a value per frequency bin or band, in Measurement.Spectrum (instruments built with NewSpectrum).
	KindSpectrum
)

// String returns the kind's name.
func (k Kind) String() string {
	switch k {
	case KindScalar:
		return "scalar"
	case KindSpectrum:
		return "spectrum"
	default:
		return "Kind(" + strconv.Itoa(int(k)) + ")"
	}
}

// Measurement is the aggregated value of one Instrument for one channel over a window of frames. A spectrum covers the
// frames of the STFT frames it aggregates, which overlap the previous measurement's when the hop is below the frame size.
type Measurement struct {
	Name       string
	Kind       Kind
	Channel    int   // 0-based channel, or Mixdown
	Start, End int64 // frames [Start, End) of the stream the value covers
	SampleRate int   // to turn frames into time
	Value      float64
	Spectrum   []Bin // KindSpectrum: ascending frequencies; allocated for every collection
}

// Bin is the value of one frequency bin, or one band, of a spectrum.
type Bin struct {
	Frequency float64 // centre, Hz: the bin's centre, or the band's nominal centre
	Low, High float64 // edges, Hz
	Value     float64 // linear amplitude: 1 is a full-scale sine (0 dBFS)
}

// Func measures one channel of a Buffer: samples holds that channel's samples, rate is the sample rate in Hz.
type Func func(samples []float64, rate int) float64

// Instrument is a named measurement and the aggregation of its values between collections. Build it with NewInstrument
// (scalar) or NewSpectrum (spectral).
type Instrument struct {
	name     string
	fn       Func               // KindScalar
	agg      func() Aggregation // KindScalar
	spectrum *spectrumConfig    // KindSpectrum
	mixdown  bool
}

// InstrumentOption configures an Instrument while NewInstrument or NewSpectrum builds it; a built Instrument never
// changes. Options that only apply to spectral instruments (WithFrameSize, WithHop, WithWindow, WithBands) make
// NewInstrument fail.
type InstrumentOption func(*instrumentOptions) error

// instrumentOptions collects the options of the Instrument being built.
type instrumentOptions struct {
	mixdown  bool
	spectrum *spectrumConfig // set by the first spectrum option
}

// OnMixdown measures the average of all channels, reported as channel Mixdown, instead of each channel (the default,
// CLAUDE.md §7 Q9).
func OnMixdown() InstrumentOption {
	return func(o *instrumentOptions) error {
		o.mixdown = true

		return nil
	}
}

// NewInstrument returns a scalar Instrument named name that measures fn and aggregates with agg (for example Max).
func NewInstrument(name string, fn Func, agg func() Aggregation, opts ...InstrumentOption) (Instrument, error) {
	if name == "" || fn == nil || agg == nil {
		return Instrument{}, fmt.Errorf("%w: name %q, func %t, aggregation %t", ErrInvalidInstrument, name, fn != nil,
			agg != nil)
	}

	var o instrumentOptions

	if err := o.apply(name, opts); err != nil {
		return Instrument{}, err
	}

	if o.spectrum != nil {
		return Instrument{}, fmt.Errorf("%w: %q: a spectrum option on a scalar instrument", ErrInvalidInstrument, name)
	}

	return Instrument{name: name, fn: fn, agg: agg, mixdown: o.mixdown}, nil
}

func (o *instrumentOptions) apply(name string, opts []InstrumentOption) error {
	for _, opt := range opts {
		if opt == nil {
			continue
		}

		if err := opt(o); err != nil {
			return fmt.Errorf("%w: %q: %w", ErrInvalidInstrument, name, err)
		}
	}

	return nil
}

// Name returns the instrument's name.
func (i Instrument) Name() string { return i.name }

// Kind returns the kind of the instrument's measurements.
func (i Instrument) Kind() Kind {
	if i.spectrum != nil {
		return KindSpectrum
	}

	return KindScalar
}

func (i Instrument) valid() bool {
	return i.name != "" && ((i.fn != nil && i.agg != nil) != (i.spectrum != nil && i.spectrum.agg != nil))
}

// Levels returns the standard level instruments, measured per channel: the peak (max), RMS (mean), crest factor (max),
// DC offset (mean), zero-crossing rate (mean) and clipped samples (sum, at full scale).
func Levels() []Instrument {
	must := func(i Instrument, err error) Instrument {
		if err != nil {
			panic(err) // the arguments are constants
		}

		return i
	}

	const (
		bits16 = 16
		clipAt = 1 - 1.0/(1<<(bits16-1)) // the largest positive 16-bit sample; wider PCM clips above it too
	)

	return []Instrument{
		must(NewInstrument("audio.peak", func(x []float64, _ int) float64 { return feature.Peak(x) }, Max)),
		must(NewInstrument("audio.rms", func(x []float64, _ int) float64 { return feature.RMS(x) }, Mean)),
		must(NewInstrument("audio.crest_factor", func(x []float64, _ int) float64 { return feature.Crest(x) }, Max)),
		must(NewInstrument("audio.dc_offset", func(x []float64, _ int) float64 { return feature.DCOffset(x) }, Mean)),
		must(NewInstrument("audio.zero_crossing_rate", feature.ZeroCrossingRate, Mean)),
		must(NewInstrument("audio.clipped_samples", func(x []float64, _ int) float64 {
			return float64(feature.Clipped(x, clipAt))
		}, Sum)),
	}
}

// key identifies a series: one instrument on one channel.
type key struct {
	instrument int
	channel    int
}

// series is one instrument on one channel. It holds the windows recorded since the previous collection: a stream that
// restarts its frame numbering (a reconnect) or changes sample rate opens a new window, so that every Measurement
// covers one coherent stretch of one stream. Spectral series also carry their STFT state across buffers.
type series struct {
	windows  []*window
	spectral *spectralState // KindSpectrum
}

type window struct {
	agg        Aggregation         // KindScalar
	spectrum   SpectrumAggregation // KindSpectrum
	layout     []spectralBin       // KindSpectrum: the frequencies of the aggregated bins
	start, end int64               // the frames of the buffers recorded
	analysed   [2]int64            // KindSpectrum: the frames [from, to) of the STFT frames recorded
	rate       int
}

// Meter measures every Buffer it receives with its Instruments. It implements sdk.Exporter.
type Meter struct {
	instruments []Instrument
	reader      Reader

	mu      sync.Mutex
	series  map[key]*series
	order   []key // sorted by instrument, then channel: the collection order
	layouts map[layoutKey]*spectralLayout
	scratch []float64 // the channel being measured
	levels  []float64 // the band levels of the frame being recorded
	closed  bool

	shutdownOnce sync.Once
	shutdownDone chan struct{}
	shutdownErr  error
}

// NewMeter returns a Meter whose measurements reader collects.
func NewMeter(reader Reader, instruments ...Instrument) (*Meter, error) {
	if reader == nil {
		return nil, fmt.Errorf("%w: nil reader", sdk.ErrInvalidOption)
	}

	names := make(map[string]bool, len(instruments))

	for _, i := range instruments {
		switch {
		case !i.valid():
			return nil, fmt.Errorf("%w: %q (build instruments with NewInstrument or NewSpectrum)", ErrInvalidInstrument,
				i.name)
		case names[i.name]:
			return nil, fmt.Errorf("%w: %q: duplicate name", ErrInvalidInstrument, i.name)
		}

		names[i.name] = true
	}

	m := &Meter{
		instruments:  append([]Instrument(nil), instruments...),
		reader:       reader,
		series:       make(map[key]*series),
		layouts:      make(map[layoutKey]*spectralLayout),
		shutdownDone: make(chan struct{}),
	}

	if err := reader.register(m.collect); err != nil {
		return nil, err
	}

	return m, nil
}

// Export implements sdk.Exporter: it measures b and records the values. It never keeps b.
func (m *Meter) Export(_ context.Context, b audio.Buffer) error {
	frames := b.Frames()
	if frames == 0 {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return sdk.ErrShutdown
	}

	for idx, inst := range m.instruments {
		if inst.mixdown {
			m.scratch = b.Mixdown(m.scratch[:0])
			m.record(key{idx, Mixdown}, inst, b, m.scratch)

			continue
		}

		for ch := range b.Format.Channels {
			m.scratch = b.Channel(ch, m.scratch[:0])
			m.record(key{idx, ch}, inst, b, m.scratch)
		}
	}

	return nil
}

func (m *Meter) record(k key, inst Instrument, b audio.Buffer, samples []float64) {
	s, ok := m.series[k]
	if !ok {
		s = &series{}
		m.series[k] = s

		i, _ := slices.BinarySearchFunc(m.order, k, compareKeys)
		m.order = slices.Insert(m.order, i, k)
	}

	frames, rate := b.Frames(), b.Format.SampleRate

	if inst.spectrum != nil {
		m.recordSpectrum(k, s, inst.spectrum, b, samples)

		return
	}

	w := s.window(b.Frame, rate, func() *window { return &window{agg: inst.agg()} })
	w.end = b.Frame + int64(frames)
	w.agg.Record(inst.fn(samples, rate), frames)
}

// window returns the window that records a buffer starting at frame, at rate Hz, opening one with open when needed:
// frames going backwards (a restarted stream) or a new rate cannot extend the last window; a gap (dropped buffers) can.
func (s *series) window(frame int64, rate int, open func() *window) *window {
	if n := len(s.windows); n > 0 {
		if w := s.windows[n-1]; frame >= w.end && rate == w.rate {
			return w
		}
	}

	w := open()
	w.start, w.end, w.rate = frame, frame, rate
	s.windows = append(s.windows, w)

	return w
}

func compareKeys(a, b key) int {
	if a.instrument != b.instrument {
		return a.instrument - b.instrument
	}

	return a.channel - b.channel
}

// collect returns the measurements recorded since the previous call, ordered by instrument, then channel (then window),
// and resets the series.
func (m *Meter) collect() []Measurement {
	m.mu.Lock()
	defer m.mu.Unlock()

	out := make([]Measurement, 0, len(m.order))

	for _, k := range m.order {
		s := m.series[k]

		for _, w := range s.windows {
			ms := Measurement{
				Name:       m.instruments[k.instrument].name,
				Channel:    k.channel,
				Start:      w.start,
				End:        w.end,
				SampleRate: w.rate,
			}

			var ok bool

			if w.spectrum != nil {
				ms.Kind, ms.Start, ms.End = KindSpectrum, w.analysed[0], w.analysed[1]
				ms.Spectrum, ok = w.collectSpectrum()
			} else {
				ms.Value, ok = w.agg.Collect()
			}

			if ok {
				out = append(out, ms)
			}
		}

		s.windows = s.windows[:0]
	}

	return out
}

// ForceFlush implements sdk.Exporter: the reader collects and emits now.
func (m *Meter) ForceFlush(ctx context.Context) error {
	m.mu.Lock()
	closed := m.closed
	m.mu.Unlock()

	if closed {
		return sdk.ErrShutdown
	}

	return m.reader.ForceFlush(ctx)
}

// Shutdown implements sdk.Exporter: it stops recording and shuts the reader down. A PeriodicReader emits one last
// batch; with a ManualReader, whatever was recorded since the last Collect is dropped, as nothing pulls it any more.
// Shutdown is idempotent and safe for concurrent use: every call waits for the same shutdown up to its own ctx, which
// runs with the first caller's ctx.
func (m *Meter) Shutdown(ctx context.Context) error {
	m.shutdownOnce.Do(func() {
		m.mu.Lock()
		m.closed = true
		m.mu.Unlock()

		go func() {
			m.shutdownErr = m.reader.Shutdown(ctx)
			close(m.shutdownDone)
		}()
	})

	return awaitResult(ctx, m.shutdownDone, func() error { return m.shutdownErr })
}

// awaitResult waits for done and returns *result, or ctx's error if ctx ends first; an available result wins.
func awaitResult(ctx context.Context, done <-chan struct{}, result func() error) error {
	select {
	case <-done:
		return result()
	default:
	}

	select {
	case <-done:
		return result()
	case <-ctx.Done():
		return ctx.Err()
	}
}
