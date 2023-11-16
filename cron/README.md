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

_TBD_

_______

### Structure and observations

_TBD_

_______

### Example

_TBD_

Another working example is the [Steam CLI app]() mentioned in the [Motivation](#motivation) section above. This 
application exposes some commands, one of them being 
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