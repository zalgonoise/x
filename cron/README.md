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
flow of job selection and execution. The runtime will allow cron to be executed as a goroutine, as its 
[`Runtime.Run`](./cron.go#L60) method has no returns, and errors are channeled via its [`Runtime.Err`](./cron.go#L77) 
method (which returns an error channel). The actual runtime of the cron is still managed with a `context.Context` that 
is provided when calling [`Runtime.Run`](./cron.go#L60) -- which can impose a cancellation or timeout strategy.

Just like the simple example above, creating a cron runtime starts with the 
[`cron.New` constructor function](./cron.go#L86).

This function only has [a variadic parameter for `cfg.Option[cron.Config]`](./cron.go#L86). This allows full modularity
on the way you build your cron runtime, to be as simple or as detailed as you want it to be -- provided that it complies 
with the minimum requirements to create one; to supply either:
- a [`selector.Selector`](./selector/selector.go#L36) 
- or, a (set of) [`executor.Runner`](./executor/executor.go#L40). This can be supplied as 
[`executor.Runnable`](./executor/executor.go#L53) as well.

```go
func New(options ...cfg.Option[Config]) (Runtime, error)
```

Below is a table with all the options available for creating a cron runtime:

|                   Function                    |                                       Input Parameters                                       |                                                                                               Description                                                                                               |
|:---------------------------------------------:|:--------------------------------------------------------------------------------------------:|:-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------:|
|    [`WithSelector`](./cron_config.go#L31)     |                    [`sel selector.Selector`](./selector/selector.go#L36)                     |                                               Configures the [`Runtime`](./cron.go#L33) with the input [`selector.Selector`](./selector/selector.go#L36).                                               |
|       [`WithJob`](./cron_config.go#L53)       | `id string`, `cronString string`, [`runners ...executor.Runner`](./executor/executor.go#L40) | Adds a new [`executor.Executor`](./executor/executor.go#L84) to the [`Runtime`](./cron.go#L33) configuration from the input ID, cron string and set of [`executor.Runner`](./executor/executor.go#L40). |
| [`WithErrorBufferSize`](./cron_config.go#L83) |                                          `size int`                                          |                                   Defines the capacity of the error channel that the [`Runtime`](./cron.go#L33) exposes in its [`Runtime.Err`](./cron.go#L77) method.                                   |
|     [`WithMetrics`](./cron_config.go#L96)     |                        [`m cron.Metrics`](./cron_with_metrics.go#L10)                        |                                                                Decorates the [`Runtime`](./cron.go#L33) with the input metrics registry.                                                                |
|     [`WithLogger`](./cron_config.go#L109)     |                 [`logger *slog.Logger`](https://pkg.go.dev/log/slog#Logger)                  |                                                                     Decorates the [`Runtime`](./cron.go#L33) with the input logger.                                                                     |
|   [`WithLogHandler`](./cron_config.go#L122)   |                [`handler slog.Handler`](https://pkg.go.dev/log/slog#Handler)                 |                                                           Decorates the [`Runtime`](./cron.go#L33) with logging using the input log handler.                                                            |
|     [`WithTrace`](./cron_config.go#L135)      |      [`tracer trace.Tracer`](https://pkg.go.dev/go.opentelemetry.io/otel/trace#Tracer)       |                                                                  Decorates the [`Runtime`](./cron.go#L33) with the input trace.Tracer.                                                                  |

The simplest possible cron runtime could be the result of a call to [`cron.New`](./cron.go#L86) with a single 
[`cron.WithJob`](./cron_config.go#L53) option. This creates all the components that a cron runtime needs with the most
minimal setup. It creates the underlying selector and executors.

Otherwise, the caller must use the [`WithSelector`](./cron_config.go#L31) option, and configure a 
[`selector.Selector`](./selector/selector.go#L36) manually when doing so. This results in more _boilerplate_ to get the
runtime set up, but provides deeper control on how the cron should be composed. The next chapter covers what is a
[`selector.Selector`](./selector/selector.go#L36) and how to create one.

_______

#### Cron Selector

This component is responsible for picking up the next job to execute, according to their schedule frequency. For this, 
the [`Selector`](./selector/selector.go#L36) is configured with a set of 
[`executor.Executor`](./executor/executor.go#L84), which in turn will expose a 
[`Next` method](./executor/executor.go#L92). With this information, the [`Selector`](./selector/selector.go#L36) cycles 
through its [`executor.Executor`](./executor/executor.go#L84) and picks up the next task(s) to run.

While the [`Selector`](./selector/selector.go#L36) calls the 
[`executor.Executor`'s `Exec` method](./executor/executor.go#L90), the actual waiting is within the
[`executor.Executor`'s](./executor/executor.go#L84) logic.

You're able to create a [`Selector`](./selector/selector.go#L36) through 
[its constructor function](./selector/selector.go#L115):

```go
func New(options ...cfg.Option[Config]) (Selector, error)
```


Below is a table with all the options available for creating a cron job selector:


|                       Function                        |                                 Input Parameters                                  |                                                         Description                                                          |
|:-----------------------------------------------------:|:---------------------------------------------------------------------------------:|:----------------------------------------------------------------------------------------------------------------------------:|
| [`WithExecutors`](./selector/selector_config.go#L23)  |          [`executors ...executor.Executor`](./executor/executor.go#L84)           | Configures the [`Selector`](./selector/selector.go#L36) with the input [`executor.Executor`(s)](./executor/executor.go#L84). |
|  [`WithMetrics`](./selector/selector_config.go#L51)   |          [`m selector.Metrics`](./selector/selector_with_metrics.go#L10)          |                   Decorates the [`Selector`](./selector/selector.go#L36) with the input metrics registry.                    |
|   [`WithLogger`](./selector/selector_config.go#L64)   |            [`logger *slog.Logger`](https://pkg.go.dev/log/slog#Logger)            |                        Decorates the [`Selector`](./selector/selector.go#L36) with the input logger.                         |
| [`WithLogHandler`](./selector/selector_config.go#L77) |           [`handler slog.Handler`](https://pkg.go.dev/log/slog#Handler)           |               Decorates the [`Selector`](./selector/selector.go#L36) with logging using the input log handler.               |
|   [`WithTrace`](./selector/selector_config.go#L90)    | [`tracer trace.Tracer`](https://pkg.go.dev/go.opentelemetry.io/otel/trace#Tracer) |                     Decorates the [`Selector`](./selector/selector.go#L36) with the input trace.Tracer.                      |

_______

#### Cron Executor

Like the name implies, the [`Executor`](./executor/executor.go#L84) is the component that actually executes the job, on 
its next scheduled time.

The [`Executor`](./executor/executor.go#L84) is composed of a [cron schedule](#cron-schedule) and a (set of) 
[`Runner`(s)](./executor/executor.go#L40). Also, the [`Executor`](./executor/executor.go#L84) stores an ID that is used 
to identify this particular job.

Having these 3 components in mind, it's natural that the [`Executor`](./executor/executor.go#L84) exposes three methods:
- [`Exec`](./executor/executor.go#L90) - runs the task when on its scheduled time.
- [`Next`](./executor/executor.go#L92) - calls the underlying 
[`schedule.Scheduler` Next method](./schedule/scheduler.go#L26).
- [`ID`](./executor/executor.go#L94) - returns the ID.

Considering that the [`Executor`](./executor/executor.go#L84) holds a specific 
[`schedule.Scheduler`](./schedule/scheduler.go#L24), it is also responsible for managing any waiting time before the 
job is executed. The strategy employed by the [`Executable`](./executor/executor.go#L99) type is one that calculates the
duration until the next job, and sleeps until that time is reached (instead of, for example, calling the
[`schedule.Scheduler` Next method](./schedule/scheduler.go#L26) every second).


To create an [`Executor`](./executor/executor.go#L84), you can use the [`New`](./executor/executor.go#L160) function 
that serves as a constructor. Note that the minimum requirements to creating an [`Executor`](./executor/executor.go#L84)
are to include both a [`schedule.Scheduler`](./schedule/scheduler.go#L24) with the 
[`WithScheduler`](./executor/executor_config.go#L60) option (or a cron string, using the 
[`WithSchedule`](./executor/executor_config.go#L77) option), 
and at least one [`Runner`](./executor/executor.go#L40) with the [`WithRunners`](./executor/executor_config.go#L28) 
option.

The [`Runner`](./executor/executor.go#L40) itself is an interface with a single method 
([`Run`](./executor/executor.go#L47)), that takes in a `context.Context` and returns an error. If your implementation is
so simple that you have it as a function and don't need to create a type for this 
[`Runner`](./executor/executor.go#L40), then you can use the [`Runnable` type](./executor/executor.go#L53) instead, 
which is a type alias to a function of the same signature, but implements [`Runner`](./executor/executor.go#L40) by 
calling itself as a function, in its `Run` method.

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