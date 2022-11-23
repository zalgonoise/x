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

__________

## Features

### Logger

The Logger is an interface that implements a Printer interface (with methods corresponding log printing actions like `Log()` and `Info()`) as well as a set of additional helper methods to make it easier to use and configure.

To spawn a Logger, you need to provide a [Handler](#handler).

```go
// Logger interface describes the behavior that a logger should
// have
//
// This includes the Printer interface, as well as other methods
// to give the logger more flexibility
type Logger interface {
	// Printer interface allows registering log messages
	Printer
	// Enabled returns a boolean on whether the logger is accepting
	// records with log level `level`
	Enabled(level level.Level) bool
	// Handler returns this Logger's Handler interface
	Handler() handlers.Handler
	// With will spawn a copy of this Logger with the input attributes
	// `attrs`
	With(attrs ...attr.Attr) Logger
}


// Printer interface describes the behavior that a (log) Printer
// should have. This includes individual methods for printing log
// messages for each log level, as well as a general-purpose `Log()`
// method to customize the log level.
type Printer interface {
	// Trace prints a log message `msg` with attributes `attrs`, with
	// Trace-level
	Trace(msg string, attrs ...attr.Attr)
	// Debug prints a log message `msg` with attributes `attrs`, with
	// Debug-level
	Debug(msg string, attrs ...attr.Attr)
	// Info prints a log message `msg` with attributes `attrs`, with
	// Info-level
	Info(msg string, attrs ...attr.Attr)
	// Warn prints a log message `msg` with attributes `attrs`, with
	// Warn-level
	Warn(msg string, attrs ...attr.Attr)
	// Error prints a log message `msg` with attributes `attrs`, with
	// Error-level
	Error(msg string, attrs ...attr.Attr)
	// Fatal prints a log message `msg` with attributes `attrs`, with
	// Fatal-level
	Fatal(msg string, attrs ...attr.Attr)
	// Log prints a log message `msg` with attributes `attrs`, with
	// `level` log level
	Log(level level.Level, msg string, attrs ...attr.Attr)
}
```


### Handler

A handler is the logging backend, responsible for writing the records with a certain format using an io.Writer. This library exposes a basic text handler that formats `any` types as string simply using `fmt.Sprintf("%v", r.Value())`; as well as a JSON handler that uses [`goccy/go-json`](https://github.com/goccy/go-json).

The text handler is not optimized for performance and is not exactly most suitable for production. The JSON handler is reliable, however, and it is safe to use in production.

The data structures implementing these handlers are immutable. The handler is a simple interface which can be implemented with the following methods:

```go
// Handler describes a logging backend, capable of writing a Record to an
// io.Writer (with its Handle() method).
//
// Beyond this feature, it also exposes methods of copying it with different
// configuration options.
type Handler interface {
	// Enabled returns a boolean on whether the Handler is accepting
	// records with log level `level`
	Enabled(level level.Level) bool
	// Handle will process the input Record, returning an error if raised
	Handle(records.Record) error
	// With will spawn a copy of this Handler with the input attributes
	// `attrs`
	With(attrs ...attr.Attr) Handler

	// WithSource will spawn a new copy of this Handler with the setting
	// to add a source file+line reference to `addSource` boolean
	WithSource(addSource bool) Handler

	// WithLevel will spawn a copy of this Handler with the input level `level`
	// as a verbosity filter
	WithLevel(level level.Level) Handler

	// WithReplaceFn will spawn a copy of this Handler with the input attribute
	// replace function `fn`
	WithReplaceFn(fn func(a attr.Attr) attr.Attr) Handler
}
```

### Record

A record is an interface exposes a set of getter methods for its elements, as well as additional helper methods to make it more granular. Although the built-in handlers already generate records themselves in their implementations, the point to the interface is to allow easy integration and extension of this library, with your own custom data types.

A record is an immutable entity.

```go
// Record interface describes the behavior that a Record should have
//
// It expose getter methods for its elements, as well as two helper methods:
//   - `AddAttr()` will return a copy of this Record with the input Attr appended
//     to the existing ones
//   - `AttrLen()` will return the length of the attributes in the record
type Record interface {
	// AddAttr returns a copy of this Record with the input Attr appended to the
	// existing ones
	AddAttr(a ...attr.Attr) Record
	// Attrs returns the slice of Attr associated to this Record
	Attrs() []attr.Attr
	// AttrLen returns the length of the slice of Attr in the Record
	AttrLen() int
	// Message returns the string Message associated to this Record
	Message() string
	// Time returns the time.Time timestamp associated to this Record
	Time() time.Time
	// Level returns the level.Level level associated to this Record
	Level() level.Level
}

```

### Attribute

An attribute is a simple interface that exposes getter and setter methods for an attribute, which can be of any type. Note that an attribute is an immutable entity.

```go
// Attr interface describes the behavior that a serializable attribute
// should have.
//
// Besides retrieving its key and value, it also permits creating a copy of
// the original Attr with a different key or a different value
type Attr interface {
	// Key returns the string key of the attribute Attr
	Key() string
	// Value returns the (any) value of the attribute Attr
	Value() any
	// WithKey returns a copy of this Attr, with key `key`
	WithKey(key string) Attr
	// WithValue returns a copy of this Attr, with value `value`
	//
	// It must be the same type of the original Attr, otherwise returns
	// nil
	WithValue(value any) Attr
}
```

Despite being exposed and used as an interface, creating a new attribute with `attr.New[T](key string, value T) attr.Attr` uses a generic function that scopes this attribute to a certain type. 

This means that when copying an attribute with the `WithValue()` method, the input value (as type `any`) must match the original attribute's type.

### Level

A level is an interface that exposes two methods, `String() string` and `Int() int`, which define different log levels in the records. While levels are used to resemble severity of the log record, they are also used by handlers (and likewise loggers) as a records filter. 

```go
// Level interface describes the behavior that a log level should have
//
// It must provide methods to be casted as a string or as an int
type Level interface {
	// String returns the level as a string
	String() string
	// Int returns the level as an int
	Int() int
}
```

### Context Logger

A logger can be embeded into a `context.Context`, and retrieved from one, too:

```go
// CtxLoggerKey is a custom type to define context keys for this
// library's logger
type CtxLoggerKey string

// StandardCtxKey is an instance of CtxLoggerKey with value "logger"
const StandardCtxKey CtxLoggerKey = "logger"

// InContext returns a copy of the input Context `ctx` with the input
// Logger `logger` as a value (identified by `StandardCtxKey`)
func InContext(ctx context.Context, logger Logger) context.Context

// From returns a Logger from the input Context `ctx`. If not present,
// it returns nil
func From(ctx context.Context) Logger
```


________________

## Disclaimer

Although `logx` isn't *the world's fastest structured logger*, I am not aiming for it either. In reality, logging should be kept simple and the right tools should be used for the job.

This means if you're concerned about metrics, setup your observability accordingly. If you need alerts on certain events, setup your observability accordingly.

All in all, logging is part of your observability strategy but it should not be the center point nor should it be the only tool in your toolbox for it. 

I love structured logging but parsing millions of lines of logs to find a single event drives one not to use it as it is inteded. Keeping logging simple and *just the right amount* is key. 

The point is, do not overburden your app with log entries as it will most certainly backfire, and you won't care about those logs. If you really need to retain big volumes of logs, surely you're using the right tool for it as well.

Getting this out of the way, let's crunch some numbers:

_________________


## Benchmarks

Setting up a quick and easy benchmark test file similar to [`zlog`'s](https://github.com/zalgonoise/zlog), in [`benchmark/benchmark_test.go`](./benchmark/benchmark_test.go):

```
# with `prettybench`:

goos: linux
goarch: amd64
pkg: github.com/zalgonoise/x/log/benchmark
cpu: AMD Ryzen 3 PRO 3300U w/ Radeon Vega Mobile Gfx
PASS
coverage: [no statements]
benchmark                                       iter       time/iter   bytes alloc         allocs
---------                                       ----       ---------   -----------         ------
BenchmarkLogger/Writing/SimpleText/LogX-4    916345   1605.00 ns/op      537 B/op    4 allocs/op
BenchmarkLogger/Writing/SimpleJSON/LogX-4    644287   1824.00 ns/op      352 B/op    5 allocs/op
BenchmarkLogger/Writing/ComplexText/LogX-4   253484   4982.00 ns/op     1271 B/op   24 allocs/op
BenchmarkLogger/Writing/ComplexJSON/LogX-4   208918   5727.00 ns/op     1544 B/op   18 allocs/op
ok      github.com/zalgonoise/x/log/benchmark   7.869s
```

When comparing these results to the vendor benchmark test in [`zlog`'s benchmarks summary](https://github.com/zalgonoise/zlog/tree/master/benchmark), it's clear that there is a major improvement when comparing to `zlog`, as well as being close to `zap` in number of allocations. Adding the results above for context, in an ordered list of tests:

```
goos: linux
goarch: amd64
pkg: github.com/zalgonoise/zlog/benchmark
cpu: AMD Ryzen 3 PRO 3300U w/ Radeon Vega Mobile Gfx

PASS
coverage: [no statements]
benchmark                                                         iter        time/iter   bytes alloc         allocs
---------                                                         ----        ---------   -----------         ------
BenchmarkVendorLoggers/Writing/SimpleText/ZeroLogger-4         2570126     471.30 ns/op      156 B/op    0 allocs/op
BenchmarkVendorLoggers/Writing/SimpleText/StdLibLogger-4       2948067     412.70 ns/op       24 B/op    1 allocs/op
BenchmarkVendorLoggers/Writing/SimpleText/ZapLogger-4           827116    1396.00 ns/op       64 B/op    3 allocs/op
BenchmarkLogger/Writing/SimpleText/LogX-4    				    916345    1605.00 ns/op      537 B/op    4 allocs/op
BenchmarkVendorLoggers/Writing/SimpleText/ZlogLogger-4          945510    1336.00 ns/op      368 B/op    9 allocs/op
BenchmarkVendorLoggers/Writing/SimpleText/LogrusLogger-4        273914    4253.00 ns/op      480 B/op   15 allocs/op

BenchmarkVendorLoggers/Writing/SimpleJSON/ZeroLogger-4         5817523     317.00 ns/op       92 B/op    0 allocs/op
BenchmarkVendorLoggers/Writing/SimpleJSON/ZapLogger-4          1000000    1152.00 ns/op        0 B/op    0 allocs/op
BenchmarkLogger/Writing/SimpleJSON/LogX-4    					644287    1824.00 ns/op      352 B/op    5 allocs/op
BenchmarkVendorLoggers/Writing/SimpleJSON/ZlogLogger-4          425534    2815.00 ns/op      376 B/op    6 allocs/op
BenchmarkVendorLoggers/Writing/SimpleJSON/LogrusLogger-4        203432    5298.00 ns/op     1080 B/op   22 allocs/op

BenchmarkVendorLoggers/Writing/ComplexText/ZeroLogger-4         382442    2906.00 ns/op      288 B/op   11 allocs/op
BenchmarkVendorLoggers/Writing/ComplexText/ZapLogger-4          171844    6609.00 ns/op      848 B/op   21 allocs/op
BenchmarkLogger/Writing/ComplexText/LogX-4   					253484    4982.00 ns/op     1271 B/op   24 allocs/op
BenchmarkVendorLoggers/Writing/ComplexText/ZlogLogger-4         121747   11129.00 ns/op     3756 B/op   50 allocs/op
BenchmarkVendorLoggers/Writing/ComplexText/LogrusLogger-4        71154   14105.00 ns/op     2168 B/op   43 allocs/op

BenchmarkVendorLoggers/Writing/ComplexJSON/ZeroLogger-4         388226    3722.00 ns/op      288 B/op   11 allocs/op
BenchmarkVendorLoggers/Writing/ComplexJSON/ZapLogger-4          231116    6320.00 ns/op      784 B/op   18 allocs/op
BenchmarkLogger/Writing/ComplexJSON/LogX-4   					208918    5727.00 ns/op     1544 B/op   18 allocs/op
BenchmarkVendorLoggers/Writing/ComplexJSON/ZlogLogger-4         115693   11486.00 ns/op     2680 B/op   40 allocs/op
BenchmarkVendorLoggers/Writing/ComplexJSON/LogrusLogger-4       116692   11029.00 ns/op     2592 B/op   44 allocs/op
```