## cron

### _a cron-scheduler library in Go_

_______

### Concept

`cron` is a Go library that allows adding cron-like scheduler(s) to Go apps, compatible with 
[Unix's cron time/date strings](https://en.wikipedia.org/wiki/Cron), to execute actions within the context of the app.

By itself, `cron` is a fantastic tool released in the mid-70's, written in C, where the user defines a specification in 
a crontab, a file listing jobs with a reference of a time/date specification and a (Unix) command to execute.

Within Go, it should provide the same set of features as the original binary, but served in a library as a 
(blocking / background) pluggable service. This means full compatibility with cron strings for scheduling, support for 
multiple executable types or functions, and support for multiple job scheduling (with different time/date 
specifications). Additionally, the library extends this functionality to support definition of seconds' frequency in 
cron-strings, context support, error returns, and full observability coverage (with added support for metrics, logs and 
traces decorators).   

_______

### Motivation

In a work environment, we see cron many times, in different scenarios. From cron installations in bare-metal Linux 
servers to one-shot containers configured in Kubernetes deployments. Some use-cases are very simple, others complex, but
the point is that is a tool used for nearly 50 years at the time of writing.

But the original tool is a C binary that executes a Unix command. If we want to explore schedulers for Go applications 
(e.g. a script executed every # hours), this means that the app needs to be compiled as a binary, and then to configure
a cron-job to execute the app in a command.

While this is fine, it raises the question -- what if I want to include it _within_ the application? This should make a 
lot of sense to you if you're a fan of SQLite like me.

There were already two libraries with different implementations, in Go:
- [`robfig/cron`](https://github.com/robfig/cron) with 12k GitHub stars
- [`go-co-op/gocron`](https://github.com/go-co-op/gocron) with 4.3k GitHub stars

Don't get me wrong -- there is nothing inherently wrong about these implementations; I went through them carefully both 
for insight and to understand what could I explore differently. A very obvious change would be a more _"modern"_ 
approach including a newer Go version (as these required Go 1.12 and 1.16 respectively); which by itself includes
`log/slog` and all the other observability-related decorators that also leverage `context.Context`.

Another more obvious exploration path would be the parser logic, as I could use my 
[generic lexer](https://github.com/zalgonoise/lex) and [generic parser](https://github.com/zalgonoise/parse) in order to 
potentially improve it.

Lastly I could try to split the cron service's components to be more configurable even in future iterations, once I had
decided on the general API for the library. There was enough ground to explore and to give it a go. :)

A personal project that I have [for a Steam CLI app](https://github.com/zalgonoise/x/tree/master/steam) is currently 
using this cron library to regularly check for discounts in the Steam Store, for certain products on a certain 
frequency, as configured by the user.

_______

### Usage

Using `cron` is as layered and modular as you want it to be. This chapter describes how to use the library effectively.

#### Getting `cron`

You're able to fetch `cron` as a Go module by importing it in your project and running `go mod tidy`:

```go
package main

import (
	"fmt"
	"context"
	
	"github.com/zalgonoise/x/cron"
	"github.com/zalgonoise/x/cron/executor"
)

func main() {
	fn := func(context.Context) error {
		fmt.Println("done!")

		return nil
	}
	
	c, err := cron.New(cron.WithJob("my-job", "* * * * *", executor.Runnable(fn)))
	// ...
}
```
_______


#### Cron Runtime

The runtime is the component that will control (like the name implies) how the module runs -- that is, controlling the 
flow of job selection and execution. The runtime will allow being executed as a goroutine, as its `Runtime.Run` method 
has no returns, and errors are channeled via its `Runtime.Err` method (which returns an error channel). The actual 
runtime of the cron is still managed with a `context.Context` that is provided when calling `Runtime.Run` -- which can 
impose a cancellation or timeout strategy.

Just like the simple example above, creating a cron runtime starts with the `cron.New` constructor function.

This function only has [a variadic parameter for `cfg.Option[cron.Config]`](./cron.go#L49). This allows full modularity
on the way you build your cron runtime, to be as simple or as detailed as you want it to be -- provided that it complies 
with the minimum requirements to create one.

```go
func New(options ...cfg.Option[Config]) (Runtime, error)
```

While the minimum requirements to create a cron runtime is to supply either a 
[`selector.Selector`](./selector/selector.go#L27) or a (set of) [`executor.Runner`](./executor/executor.go#L31). The 
latter can be supplied as an [`executor.Runnable`](./executor/executor.go#L35).

Below is a table with all the options available for creating a cron runtime:

|                   Function                    |                                       Input Parameters                                       |                                                        Description                                                         |
|:---------------------------------------------:|:--------------------------------------------------------------------------------------------:|:--------------------------------------------------------------------------------------------------------------------------:|
|    [`WithSelector`](./cron_config.go#L31)     |                    [`sel selector.Selector`](./selector/selector.go#L27)                     |                                Configures the `Runtime` with the input `selector.Selector`.                                |
|       [`WithJob`](./cron_config.go#L53)       | `id string`, `cronString string`, [`runners ...executor.Runner`](./executor/executor.go#L31) | Adds a new `executor.Executor` to the `Runtime` configuration from the input ID, cron string and set of `executor.Runner`. |
| [`WithErrorBufferSize`](./cron_config.go#L83) |                                          `size int`                                          |             Defines the capacity of the error channel that the `Runtime` exposes in its `Runtime.Err` method.              |
|     [`WithMetrics`](./cron_config.go#L96)     |                        [`m cron.Metrics`](./cron_with_metrics.go#L5)                         |                                  Decorates the `Runtime` with the input metrics registry.                                  |
|     [`WithLogger`](./cron_config.go#L109)     |                 [`logger *slog.Logger`](https://pkg.go.dev/log/slog#Logger)                  |                                       Decorates the `Runtime` with the input logger.                                       |
|   [`WithLogHandler`](./cron_config.go#L122)   |                [`handler slog.Handler`](https://pkg.go.dev/log/slog#Handler)                 |                              Decorates the Runtime with logging using the input log handler.                               |
|     [`WithTrace`](./cron_config.go#L135)      |      [`tracer trace.Tracer`](https://pkg.go.dev/go.opentelemetry.io/otel/trace#Tracer)       |                                     Decorates the Runtime with the input trace.Tracer.                                     |

The simplest possible cron runtime could be the result of a call to `cron.New` with a single `cron.WithJob` option. This
creates all the components that a cron runtime needs with the most minimal setup. It creates the underlying selector and 
executors.

_______

#### Cron Selector

_TBD_

_______

#### Cron Executor

_TBD_

_______

#### Cron Schedule

_TBD_

_______

##### Schedule Resolver

_TBD_

_______

##### Schedule Parser

_TBD_

_______

### Structure and observations

_TBD_

_______

### Example

_TBD_

Another working example is the [Steam CLI app](https://github.com/zalgonoise/x/tree/master/steam) mentioned in the 
[Motivation](#motivation) section above. This application exposes some commands, one of them being 
[`monitor`](https://github.com/zalgonoise/x/blob/master/steam/cmd/steam/monitor/monitor.go). This file provides some 
insight on how the cron service is set up from a `main.go` / script-like approach.

You can also take a look [at its 
`runner.go` file](https://github.com/zalgonoise/x/blob/master/steam/cmd/steam/monitor/runner.go), that implements the 
`executor.Runner` interface.

_______

### Disclaimer

This is not a one-size-fits-all solution! Please take your time to evaluate it for your own needs with due diligence.
While having _a library for this and a library for that_ is pretty nice, it could potentially be only overhead hindering
the true potential of your app! Be sure to read the code that you are using to be a better judge if it is a good fit for
your project. With that in mind, I hope you enjoy this library. Feel free to contribute by filing either an issue or a
pull request.