# spanner

A simple trace producer and exporter written in Go

___________________


## Overview

After working with [OpenTelemetry's Tracer](https://github.com/open-telemetry/opentelemetry-go/tree/main/trace), I decided to implement a *simpler* tracer for my own applications. This tracer would output data in the same format as OpenTelemetry's, but have a smaller surface and more concise API. This would imply mimicking some of the structure, approach and API of OpenTelemetry's Tracer, while approaching the implementation in my own way (regardless). Basically, looking at the interfaces exposed by OpenTelemetry's Tracer and implementing my own Tracer from there.

This obviously resulted in an application with 6x more allocations than the original, mostly due to the approach when writing Span data. This was a perfect occasion to dive deeper into profiling Go applications using tools like `pprof`, and gradually improving the performance of my implementation. By the time the repo the repo was ported from `zalgonoise/x/spanner` to `zalgonoise/spanner`, it had halved the amount of allocations as compared to the first implementation (so, still 3x more allocations than OpenTelemetry's).

## Installation 


To fetch `spanner` as a Go library, use `go get` or `go install`:

```
go get -u github.com/zalgonoise/spanner
```

```
go install github.com/zalgonoise/spanner@latest
```

...or, simply import it in your Go file and run `go mod tidy`:

```go
import (
    // (...)

    "github.com/zalgonoise/spanner"
)
```

## Features 

### Trace

```go
type Trace interface {
	// ID returns the TraceID
	ID() TraceID
	// Register sets the input pointer to a SpanID `s` as this Trace's reference parent_id
	Register(s *SpanID)
	// Parent returns the parent SpanID, or nil if unset
	Parent() *SpanID
}
```

### Span

```go
type Span interface {
	// Start sets the span to record
	Start()
	// End stops the span, returning the collected SpanData in the action
	End()
	// ID returns the SpanID of the Span
	ID() SpanID
	// IsRecording returns a boolean on whether the Span is currently recording
	IsRecording() bool
	// SetName overwrites the Span's name field with the string `name`
	SetName(name string)
	// SetParent overwrites the Span's parent_id field with the SpanID `id`
	SetParent(span Span)
	// Add appends attributes (key-value pairs) to the Span
	Add(attrs ...attr.Attr)
	// Attrs returns the Span's stored attributes
	Attrs() []attr.Attr
	// Replace will flush the Span's attributes and store the input attributes `attrs` in place
	Replace(attrs ...attr.Attr)
	// Event creates a new event within the Span
	Event(name string, attrs ...attr.Attr)
	// Extract returns the current SpanData for the Span, regardless of its status
	Extract() SpanData
	// Events returns the events in the Span
	Events() []EventData
}
```

#### Span Attributes

#### Span Events

### Tracer

```go
type Tracer interface {
	// Start reuses the Trace in the input context `ctx`, or creates one if it doesn't exist. It also
	// creates the Span for the action, with string name `name`. Each call creates a new Span.
	//
	// After calling Start, the input context will still reference the parent Span's ID, nil if it's a new Trace.
	// The returned context will reference the returned Span's ID, to be used as the next call's parent.
	//
	// The returned Span is required, even if to defer its closure, with `defer s.End()`. The caller MUST close the
	// returned Span.
	Start(ctx context.Context, name string) (context.Context, Span)
	// To sets the Span exporter to Exporter `e`
	To(e Exporter)
}
```

### Processor


```go
type SpanProcessor interface {
	// Handle routes the input Span `span` to the SpanProcessor's Exporter
	Handle(span Span)
	// Shutdown gracefully stops the SpanProcessor, returning an error
	Shutdown(ctx context.Context) error
	// Flush will force-push the existing SpanData in the SpanProcessor's batch into the
	// Exporter, even if not yet scheduled to do so
	Flush(ctx context.Context) error
}
```
### Exporter

```go
type Exporter interface {
	// Export pushes the input SpanData `spans` to its output, as a non-blocking
	// function
	Export(ctx context.Context, spans []SpanData) error
	// Shutdown gracefully terminates the Exporter
	Shutdown(ctx context.Context) error
}
```

### ID Generator

```go
type IDGenerator interface {
	// NewTraceID creates a new TraceID
	NewTraceID() TraceID
	// NewSpanID creates a new SpanID
	NewSpanID() SpanID
}
```

## Disclaimer

This library does not aim to replace OpenTelemetry's Tracer implementation. It is about exploring an existing concept in order to work on improving the performance of a Go application. It just happens to be a Tracer. Don't replace your in-prod OpenTelemetry Tracer with this one.

This implementation is not more performant than OpenTelemetry's Tracer, as clarified in the Benchmark section below.

This approach tries to be a simple approach to the Span data structure exposed by OpenTelemetry's Tracer.

## Benchmarks