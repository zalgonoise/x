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

## Disclaimer

## Benchmarks