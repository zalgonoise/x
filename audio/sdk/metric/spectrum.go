package metric

import (
	"errors"
	"fmt"

	"github.com/zalgonoise/x/audio"
	"github.com/zalgonoise/x/audio/dsp/spectral"
	win "github.com/zalgonoise/x/audio/fft/window"
)

// DefaultFrameSize is the frame size of spectral instruments unless WithFrameSize says otherwise: 46 ms at 44.1 kHz,
// with bins 21.5 Hz apart.
const DefaultFrameSize = 2048

// errSpectrumOption is wrapped by the errors of invalid spectrum options.
var errSpectrumOption = errors.New("invalid spectrum option")

// spectrumConfig is the configuration of a spectral instrument; it is immutable once the instrument is built.
type spectrumConfig struct {
	agg       func() SpectrumAggregation
	size, hop int
	window    func(int) win.Window
	bands     []spectral.Band
	stft      *spectral.STFT // never fed: every series clones it, sharing its window
}

// spectrumOptions returns the spectrum configuration being built, creating it on first use (so that NewInstrument can
// reject spectrum options).
func (o *instrumentOptions) spectrumOptions() *spectrumConfig {
	if o.spectrum == nil {
		o.spectrum = &spectrumConfig{}
	}

	return o.spectrum
}

// WithFrameSize sets the number of samples in each analysed frame of a spectral instrument: a power of two in
// [spectral.MinSize, spectral.MaxSize]. Bins lie rate/n Hz apart. The default is DefaultFrameSize.
func WithFrameSize(n int) InstrumentOption {
	return func(o *instrumentOptions) error {
		if n < spectral.MinSize || n > spectral.MaxSize || n&(n-1) != 0 {
			return fmt.Errorf("%w: frame size %d (want a power of two in [%d, %d])", errSpectrumOption, n,
				spectral.MinSize, spectral.MaxSize)
		}

		o.spectrumOptions().size = n

		return nil
	}
}

// WithHop sets the distance, in samples, between the starts of two analysed frames of a spectral instrument. The default
// is half the frame size (50% overlap); a hop larger than the frame size skips the samples in between.
func WithHop(n int) InstrumentOption {
	return func(o *instrumentOptions) error {
		if n < 1 {
			return fmt.Errorf("%w: hop %d", errSpectrumOption, n)
		}

		o.spectrumOptions().hop = n

		return nil
	}
}

// WithWindow sets the window of a spectral instrument: a function returning an n-point window, such as window.Hann (the
// default) or another fft/window table.
func WithWindow(fn func(n int) win.Window) InstrumentOption {
	return func(o *instrumentOptions) error {
		if fn == nil {
			return fmt.Errorf("%w: nil window", errSpectrumOption)
		}

		o.spectrumOptions().window = fn

		return nil
	}
}

// WithBands reduces each spectrum to the levels of bands (spectral.OctaveBands, for example), which must be ascending
// and must not overlap. Bands that hold no bin at the stream's resolution are left out. Without it, measurements hold
// every bin, from 0 Hz to the Nyquist frequency.
func WithBands(bands []spectral.Band) InstrumentOption {
	return func(o *instrumentOptions) error {
		if len(bands) == 0 {
			return fmt.Errorf("%w: no bands", errSpectrumOption)
		}

		o.spectrumOptions().bands = append([]spectral.Band(nil), bands...)

		return nil
	}
}

// NewSpectrum returns a spectral Instrument named name: per channel, it frames the stream with a short-time Fourier
// transform (spectral.STFT) carried across buffers, records the amplitude spectrum of every frame, or its band levels
// (WithBands), into agg (PeakHold or MeanSpectrum), and reports KindSpectrum measurements. A discontinuity in the stream
// (a restart, a gap of dropped buffers, a new sample rate) starts a new frame, and frames still incomplete when the
// stream ends are not analysed. Buffers with no sample rate are ignored.
func NewSpectrum(name string, agg func() SpectrumAggregation, opts ...InstrumentOption) (Instrument, error) {
	if name == "" || agg == nil {
		return Instrument{}, fmt.Errorf("%w: name %q, aggregation %t", ErrInvalidInstrument, name, agg != nil)
	}

	var o instrumentOptions

	cfg := o.spectrumOptions() // a fresh configuration: the built Instrument is the only holder of it

	if err := o.apply(name, opts); err != nil {
		return Instrument{}, err
	}

	cfg.agg = agg

	if cfg.size == 0 {
		cfg.size = DefaultFrameSize
	}

	if cfg.hop == 0 {
		cfg.hop = cfg.size / 2 //nolint:mnd // half the frame: 50% overlap
	}

	if cfg.window == nil {
		cfg.window = win.Hann
	}

	stft, err := spectral.NewSTFT(cfg.size, cfg.hop, cfg.window(cfg.size))
	if err != nil {
		return Instrument{}, fmt.Errorf("%w: %q: %w", ErrInvalidInstrument, name, err)
	}

	cfg.stft = stft

	if cfg.bands != nil {
		// any rate checks the order of the bands
		if _, err := spectral.NewBanding(cfg.bands, cfg.size, 1, stft.ENBW()); err != nil {
			return Instrument{}, fmt.Errorf("%w: %q: %w", ErrInvalidInstrument, name, err)
		}
	}

	return Instrument{name: name, spectrum: cfg, mixdown: o.mixdown}, nil
}

// Spectra returns the standard spectral instruments, measured per channel: "audio.spectrum.peak" (PeakHold) and
// "audio.spectrum.mean" (MeanSpectrum), both configured by opts (WithBands, WithFrameSize...).
func Spectra(opts ...InstrumentOption) ([]Instrument, error) {
	peak, err := NewSpectrum("audio.spectrum.peak", PeakHold, opts...)
	if err != nil {
		return nil, err
	}

	mean, err := NewSpectrum("audio.spectrum.mean", MeanSpectrum, opts...)
	if err != nil {
		return nil, err
	}

	return []Instrument{peak, mean}, nil
}

// spectralState is a spectral series' analysis state, carried across buffers and windows.
type spectralState struct {
	stft   *spectral.STFT
	layout *spectralLayout // at rate
	rate   int
	origin int64 // the stream frame at the STFT's position 0: its last reset
	next   int64 // the frame after the last one recorded
}

// spectralLayout maps one instrument's spectra onto frequencies at one sample rate. It never changes once built, so
// every channel of the instrument shares it (see Meter.layout).
type spectralLayout struct {
	banding *spectral.Banding // nil without bands
	bins    []spectralBin     // the frequencies of the recorded bins
}

// spectralBin is the frequency range of one recorded bin.
type spectralBin struct {
	centre, low, high float64
}

// maxLayouts bounds the layouts a Meter caches.
const maxLayouts = 64

// layoutKey identifies a spectralLayout: one instrument at one sample rate.
type layoutKey struct {
	instrument int
	rate       int
}

func (m *Meter) recordSpectrum(k key, s *series, cfg *spectrumConfig, b audio.Buffer, samples []float64) {
	rate := b.Format.SampleRate
	if rate <= 0 {
		return
	}

	st := s.spectral
	if st == nil {
		st = &spectralState{stft: cfg.stft.Clone()}
		s.spectral = st
	}

	if b.Frame != st.next || rate != st.rate {
		st.stft.Reset() // no frame straddles a discontinuity
		st.origin = b.Frame
	}

	if rate != st.rate {
		lay, ok := m.layout(k.instrument, cfg, rate)
		if !ok {
			st.layout, st.rate = nil, 0

			return
		}

		st.layout, st.rate = lay, rate
	}

	st.next = b.Frame + int64(b.Frames())

	w := s.window(b.Frame, rate, func() *window { return &window{spectrum: cfg.agg(), layout: st.layout.bins} })
	w.end = st.next

	if len(st.layout.bins) == 0 {
		return // no band holds a bin at this rate
	}

	st.stft.Write(samples, func(spectrum []float64) {
		if st.layout.banding != nil {
			m.levels = st.layout.banding.Levels(spectrum, m.levels[:0])
			spectrum = m.levels
		}

		// the STFT frame may have started in an earlier buffer, before the window (and before a collection)
		from := st.origin + st.stft.Position()
		if w.analysed == ([2]int64{}) {
			w.analysed[0] = from
		}

		w.analysed[1] = from + int64(cfg.size)
		w.spectrum.Record(spectrum, cfg.hop)
	})
}

// layout returns the mapping of an instrument's spectra onto frequencies at rate Hz, building it on first use. It
// reports false if the rate cannot be analysed.
func (m *Meter) layout(instrument int, cfg *spectrumConfig, rate int) (*spectralLayout, bool) {
	k := layoutKey{instrument: instrument, rate: rate}
	if lay, ok := m.layouts[k]; ok {
		return lay, true
	}

	if len(m.layouts) >= maxLayouts {
		clear(m.layouts) // a stream that keeps changing rate; windows keep the layouts they use
	}

	lay := &spectralLayout{}

	if cfg.bands != nil {
		banding, err := spectral.NewBanding(cfg.bands, cfg.size, rate, cfg.stft.ENBW())
		if err != nil {
			return nil, false
		}

		lay.banding = banding
		lay.bins = make([]spectralBin, 0, len(banding.Bands()))

		for _, band := range banding.Bands() {
			lay.bins = append(lay.bins, spectralBin{centre: band.Centre, low: band.Low, high: band.High})
		}
	} else {
		bins := cfg.size/2 + 1                                       //nolint:mnd // a one-sided spectrum: 0 Hz to Nyquist
		half := spectral.BinFrequency(1, cfg.size, rate) / 2         //nolint:mnd // a bin spans half a bin either side
		nyquist := spectral.BinFrequency(cfg.size/2, cfg.size, rate) //nolint:mnd // the last bin

		lay.bins = make([]spectralBin, 0, bins)

		for i := range bins {
			f := spectral.BinFrequency(i, cfg.size, rate)
			lay.bins = append(lay.bins, spectralBin{centre: f, low: max(0, f-half), high: min(nyquist, f+half)})
		}
	}

	m.layouts[k] = lay

	return lay, true
}

// collectSpectrum returns the window's aggregated spectrum, or false if no frame completed in it.
func (w *window) collectSpectrum() ([]Bin, bool) {
	v, ok := w.spectrum.Collect()
	if !ok || len(v) == 0 || len(v) != len(w.layout) {
		return nil, false
	}

	bins := make([]Bin, len(v))

	for i, b := range w.layout {
		bins[i] = Bin{Frequency: b.centre, Low: b.low, High: b.high, Value: v[i]}
	}

	return bins, true
}
