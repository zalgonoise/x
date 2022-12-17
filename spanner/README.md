# spanner

A simple trace producer and exporter written in Go

___________________


## Overview

After working with [OpenTelemetry's Tracer](https://github.com/open-telemetry/opentelemetry-go/tree/main/trace), I decided to implement a *simpler* tracer for my own applications. This tracer would output data in the same format as OpenTelemetry's, but have a smaller surface and more concise API. This would imply mimicking some of the structure, approach and API of OpenTelemetry's Tracer, while approaching the implementation in my own way (regardless). Basically, looking at the interfaces exposed by OpenTelemetry's Tracer and implementing my own Tracer from there.

This obviously resulted in an application with 6x more allocations than the original, mostly due to the approach when writing Span data. This was a perfect occasion to dive deeper into profiling Go applications using tools like `pprof`, and gradually improving the performance of my implementation. By the time the repo the repo was ported from `zalgonoise/x/spanner` to `zalgonoise/spanner`, it had decimated the amount of allocations as compared to the first implementation (however, still 25% more allocations than OpenTelemetry's). More information on benchmarks in its own section, below.

The trace output will have the same elements as found in OpenTelemetry's implementation, with a bit less metadata. Still, the overall structure described in [OpenTelemetry's Traces reference documentation](https://opentelemetry.io/docs/concepts/signals/traces/) is preserved, as seen in the example shared in it:

```json
{
    "name": "Hello-Greetings",
    "context": {
        "trace_id": "0x5b8aa5a2d2c872e8321cf37308d69df2",
        "span_id": "0x5fb397be34d26b51",
    },
    "parent_id": "0x051581bf3cb55c13",
    "start_time": "2022-04-29T18:52:58.114304Z",
    "end_time": "2022-04-29T18:52:58.114435Z",
    "attributes": {
        "http.route": "some_route1"
    },
    "events": [
        {
            "name": "hey there!",
            "timestamp": "2022-04-29T18:52:58.114561Z",
            "attributes": {
                "event_attributes": 1
            }
        },
        {
            "name": "bye now!",
            "timestamp": "2022-04-29T22:52:58.114561Z",
            "attributes": {
                "event_attributes": 1
            }
        }
    ],
}
```

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

A Trace represents a single transaction in a system. It has a unique ID (16-byte-long, hex-encoded) that is present in all Spans spawned within this Trace. It also registers a Span's unique ID for reference when creating a new Span; as it will point to its parent Span.

Traces are created when the `Tracer.Start()` method is called, and the input context does not have a Trace already.

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

A Span represents a single action within a transaction, in a system. It has a unique ID (8-byte-long, hex-encoded) and keeps track of the parent Span's ID, `nil` if it's the root Span.

While `Tracer.Start()` kicks off a Span's beginning, it is the responsibility of the caller to end it with its `Span.End()` method, deferred if needed be.

A Span may also store metadata besides a name and beginning / end timestamps. As covered below, a Span is also able to store key-value-pair attributes and events.

Lastly, a Span exposes methods of extracting both its SpanData and its EventData.

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

A Span stores key-value-pair attributes as metadata to the action in question. These key-value-pair attributes have a `string`-type key and `any`-type value, leveraging the [`zalgonoise/attr`](https://github.com/zalgonoise/attr) library.

#### Span Events

A Span also records events, which are one-shot entries with a `string` *name* and optionally any amount of attributes, leveraging the [`zalgonoise/attr`](https://github.com/zalgonoise/attr) library.

Span Events will store the Span Event name, attributes and also a timestamp of when the Event was recorded.

### Tracer

A Tracer creates a Traces and Spans within the input context, as well as setting Exporters to write the output SpanData.

It's main method `Tracer.Start()` reads the input context to create a Trace if it does not exist which is then stored in the context.

It also spawns a new Span with the input `string` *name*, that is returned to the caller.

The returned context will store the returned Span's ID, so that it is referenced when creating a new Span as a child. The input context, however, will not store the returned Span's ID, and when creating a new Span it will keep the previous parent Span's ID -- making it seem like the next call was done side-by-side with the parent (and not a child of it).

Its `Tracer.To()` method sets the Span exporter to the input Exporter.

Its `Tracer.Processor()` method returns the configured SpanProcessor.

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
	// Processor returns the configured SpanProcessor in the Tracer
	Processor() SpanProcessor
}
```

### Processor

A SpanProcessor will ingest the ended Spans when their `Span.End()` method is called, extract their SpanData, and push batches of SpanData to the configured Exporter. The SpanProcessor will be responsible of any post-recording processing that the Span needs, and it runs in a go routine.

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

An Exporter will write a batch of SpanData to a certain output, as implemented in its `Exporter.Export()` method.

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

An ID Generator creates both Trace IDs and Span IDs, using a `crypto/rand` RNG. 

TraceIDs are 16-byte-long, hex-encoded values that implement the `fmt.Stringer` interface.

SpanIDs are 8-byte-long, hex-encoded values that implement the `fmt.Stringer` interface.

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

> Tests are piped to [`prettybench`](https://github.com/cespare/prettybench) for a cleaner output of benchmark results.

To benchmark this implementation in comparison to OpenTelemetry's, the best route is to follow the [OpenTelemetry Go's Getting Started documentation](https://opentelemetry.io/docs/instrumentation/go/getting-started/), that describes a reasonable Fibonacci application with 3 different modules, that can be traced individually. As such, there are three implementations of that same logic in [`zalgonoise/x/benchmark/spanner`](https://github.com/zalgonoise/x/tree/master/benchmark/spanner):

- [`zalgonoise/x/benchmark/spanner/core`](https://github.com/zalgonoise/x/tree/master/benchmark/spanner/core): contains the raw logic of the Fibonacci application, to benchmark with no Tracers involved.
- [`zalgonoise/x/benchmark/spanner/optl`](https://github.com/zalgonoise/x/tree/master/benchmark/spanner/optl): contains the logic of the Fibonacci application wrapped with the OpenTelemetry tracer, exporting the Spans to standard-out.
- [`zalgonoise/x/benchmark/spanner/self`](https://github.com/zalgonoise/x/tree/master/benchmark/spanner/self): contains the logic of the Fibonacci application wrapped with this tracer, exporting the Spans to standard-out. 

#### Core

For reference, note the benchmark output of the [`core`](https://github.com/zalgonoise/x/blob/master/benchmark/spanner/core/core_test.go#L19) implementation:

```
❯ go test -bench . -benchtime=10s -benchmem -cpuprofile /tmp/cpu.pprof -run BenchmarkRuntime | prettybench

goos: linux
goarch: amd64
pkg: github.com/zalgonoise/x/benchmark/spanner/core
cpu: AMD Ryzen 3 PRO 3300U w/ Radeon Vega Mobile Gfx
PASS
benchmark                iter      time/iter   bytes alloc        allocs
---------                ----      ---------   -----------        ------
BenchmarkRuntime-4   36876082   421.60 ns/op        8 B/op   1 allocs/op
ok      github.com/zalgonoise/x/benchmark/spanner/core  16.064s
```


#### OpenTelemetry

OpenTelemetry's test instruments the application just like the official Getting Started guide suggests, but also involves flushing the accumulated Span data to the official standard-out exporter:

```
❯ go test -bench . -benchtime=10s -benchmem -cpuprofile /tmp/cpu.pprof -run BenchmarkRuntime | prettybench
goos: linux
goarch: amd64
pkg: github.com/zalgonoise/x/benchmark/spanner/optl
cpu: AMD Ryzen 3 PRO 3300U w/ Radeon Vega Mobile Gfx
PASS
benchmark              iter      time/iter   bytes alloc         allocs
---------              ----      ---------   -----------         ------
BenchmarkRuntime-4   283874    42.95 μs/op    12405 B/op   75 allocs/op
ok      github.com/zalgonoise/x/benchmark/spanner/optl  22.038s
```

#### `zalgonoise/spanner`

This repo's test instruments the application exactly the same way as OpenTelemetry, also exporting the Span data to standard-out:

```
❯ go test -bench . -benchtime=10s -benchmem -cpuprofile /tmp/cpu.pprof -run BenchmarkRuntime | prettybench

goos: linux
goarch: amd64
pkg: github.com/zalgonoise/x/benchmark/spanner/self
cpu: AMD Ryzen 3 PRO 3300U w/ Radeon Vega Mobile Gfx
PASS
benchmark              iter      time/iter   bytes alloc         allocs
---------              ----      ---------   -----------         ------
BenchmarkRuntime-4   169114    73.01 μs/op     6343 B/op   88 allocs/op
ok      github.com/zalgonoise/x/benchmark/spanner/self  13.237s
```
