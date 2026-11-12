package sdk_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	otelmetric "go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"

	"github.com/zalgonoise/x/audio/internal/testutil"
	"github.com/zalgonoise/x/audio/sdk"
)

// Test doubles built on the OTel no-op types, so the module does not depend on the OTel SDK.

type fakeCounter struct {
	metricnoop.Int64Counter

	mu  sync.Mutex
	sum int64
}

func (c *fakeCounter) Add(_ context.Context, n int64, _ ...otelmetric.AddOption) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.sum += n
}

func (c *fakeCounter) value() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.sum
}

type fakeMeterProvider struct {
	metricnoop.MeterProvider

	mu       sync.Mutex
	counters map[string]*fakeCounter
}

type fakeMeter struct {
	metricnoop.Meter

	p *fakeMeterProvider
}

func (p *fakeMeterProvider) Meter(string, ...otelmetric.MeterOption) otelmetric.Meter {
	return fakeMeter{p: p}
}

func (m fakeMeter) Int64Counter(name string, _ ...otelmetric.Int64CounterOption) (otelmetric.Int64Counter, error) {
	m.p.mu.Lock()
	defer m.p.mu.Unlock()

	if m.p.counters == nil {
		m.p.counters = map[string]*fakeCounter{}
	}

	c := &fakeCounter{}
	m.p.counters[name] = c

	return c, nil
}

func (p *fakeMeterProvider) value(name string) int64 {
	p.mu.Lock()
	c := p.counters[name]
	p.mu.Unlock()

	if c == nil {
		return -1
	}

	return c.value()
}

type fakeSpan struct {
	tracenoop.Span

	mu     sync.Mutex
	name   string
	attrs  []attribute.KeyValue
	errs   []error
	status codes.Code
	ended  bool
}

func (s *fakeSpan) SetAttributes(kv ...attribute.KeyValue) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.attrs = append(s.attrs, kv...)
}

func (s *fakeSpan) RecordError(err error, _ ...trace.EventOption) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.errs = append(s.errs, err)
}

func (s *fakeSpan) SetStatus(code codes.Code, _ string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.status = code
}

func (s *fakeSpan) End(...trace.SpanEndOption) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ended = true
}

type fakeTracerProvider struct {
	tracenoop.TracerProvider

	mu    sync.Mutex
	spans []*fakeSpan
}

type fakeTracer struct {
	tracenoop.Tracer

	p *fakeTracerProvider
}

func (p *fakeTracerProvider) Tracer(string, ...trace.TracerOption) trace.Tracer {
	return fakeTracer{p: p}
}

func (t fakeTracer) Start(ctx context.Context, name string, _ ...trace.SpanStartOption) (context.Context, trace.Span) {
	s := &fakeSpan{name: name}

	t.p.mu.Lock()
	t.p.spans = append(t.p.spans, s)
	t.p.mu.Unlock()

	return trace.ContextWithSpan(ctx, s), s
}

func TestPipelineTelemetry(t *testing.T) {
	file := testutil.WAV{Channels: 2, SampleRate: 8000, BitsPerSample: 16}.
		EncodeInts([][]int32{make([]int32, 2000), make([]int32, 2000)}) // 250 ms: 3 buffers of 100 ms

	mp := &fakeMeterProvider{}
	tp := &fakeTracerProvider{}
	boom := errors.New("boom")

	p, err := sdk.NewPipeline(bytesSource(file), []sdk.Exporter{&recorder{exportErr: boom}},
		sdk.WithMeterProvider(mp), sdk.WithTracerProvider(tp), sdk.WithErrorHandler(func(error) {}))
	require.NoError(t, err)
	require.NoError(t, p.Run(t.Context()))

	require.Equal(t, int64(3), mp.value("audio.pipeline.buffers"))
	require.Equal(t, int64(2000), mp.value("audio.pipeline.frames"))
	require.Equal(t, int64(3), mp.value("audio.pipeline.export.errors"))

	require.Len(t, tp.spans, 1)

	span := tp.spans[0]
	require.Equal(t, "audio.pipeline.run", span.name)
	require.True(t, span.ended)
	require.Contains(t, span.attrs, attribute.Int("audio.sample_rate", 8000))
	require.Contains(t, span.attrs, attribute.Int("audio.channels", 2))
	require.Equal(t, codes.Unset, span.status, "a clean end of stream is not an error")

	failing, err := sdk.NewPipeline(bytesSource([]byte("not audio at all")), nil, sdk.WithTracerProvider(tp))
	require.NoError(t, err)
	require.Error(t, failing.Run(t.Context()))
	require.Equal(t, codes.Error, tp.spans[1].status)
	require.Len(t, tp.spans[1].errs, 1)

	for name, opt := range map[string]sdk.Option{
		"TracerProvider": sdk.WithTracerProvider(nil),
		"MeterProvider":  sdk.WithMeterProvider(nil),
	} {
		_, err := sdk.NewPipeline(bytesSource(nil), nil, opt)
		require.ErrorIs(t, err, sdk.ErrInvalidOption, name)
	}
}

func TestQueueTelemetry(t *testing.T) {
	mp := &fakeMeterProvider{}
	rec := &recorder{gate: make(chan struct{})}

	q, err := sdk.NewQueuedExporter(rec, sdk.WithQueueSize(1), sdk.WithQueueMeterProvider(mp))
	require.NoError(t, err)

	for i := range 4 {
		require.NoError(t, q.Export(t.Context(), buf(i)))
	}

	close(rec.gate)
	require.NoError(t, q.Shutdown(t.Context()))
	require.Equal(t, int64(q.Dropped()), mp.value("audio.queue.dropped"))
	require.Positive(t, q.Dropped())

	_, err = sdk.NewQueuedExporter(rec, sdk.WithQueueMeterProvider(nil))
	require.ErrorIs(t, err, sdk.ErrInvalidOption)
}
