# logx

*A blazing fast structured logger for Go*

___________


## Overview

After working on [`zlog`](https://github.com/zalgonoise/zlog), I've decided to do a second iteration of a structured logger with a simpler (but more meaningful) API, as well as a more performant solution.

As for the logger API, I followed most of the input shared in the discussion in [go#54763](https://github.com/golang/go/discussions/54763), on what I saw was useful and idiomatic. As for the implementation itself, although it is not as clear-cut as desired or as performant as [`zerolog`](https://github.com/rs/zerolog) or [`zap`](https://github.com/uber-go/zap), it is still going for a very low number of allocations for the amount of time put into. More information in the [benchmarks section](#benchmarks)

___________

## Installation

To fetch `logx` as a Go library, use `go get` or `go install`:

```
go get -u github.com/zalgonoise/logx
```

```
go install github.com/zalgonoise/logx@latest
```

...or, simply import it in your Go file and run `go mod tidy`:

```go
import (
    // (...)

    "github.com/zalgonoise/logx"
)

```