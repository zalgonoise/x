### `/x/` 

Experimental Go libraries _(take it with a grain of salt)_

______________

### Concept

This repository contains projects, experiments and studies when developing new Go libraries and applications.

Its content ranges from HTTP / net utilities, generics in Go, binary encoding, to audio processing and much more in between. In `/x/` many ideas are born, and some eventually ported to their own repositories.  

This is the place for trying something new, to study it and amplify an initial idea. It's a place to be unbounded and unconstrained; to both succeed and to fail.


_________

### Index 


|                Directory                 |                    Notes                    |                                                                           Description                                                                            |
|:----------------------------------------:|:-------------------------------------------:|:----------------------------------------------------------------------------------------------------------------------------------------------------------------:|
|           [`x/audio`](./audio)           |                                             |                                         digital signal processing tools for audio, including WAV encoding and decoding.                                          |
|       [`x/benchmark`](./benchmark)       |                                             |                                     several benchmark tests to compare `x` implementations with third-party implementations                                      |
|          [`x/bezier`](./bezier)          |       [🤖](https://chat.openai.com/)        |                                                                       PNG to SVG converter                                                                       |
|             [`x/bom`](./bom)             |                                             |                                                     handlers for Byte Order Mark presence in text documents                                                      |
|              [`x/cb`](./cb)              |                                             |                                                         experiments on a Circuit Breaker implementation                                                          |
|             [`x/cfg`](./cfg)             |                                             |                                                            generic configuration and options library                                                             |
|             [`x/cli`](./cli)             |                                             |                                                        Command-Line Interface utilities and helper logic                                                         |
|       [`x/codegraph`](./codegraph)       |                                             |                                            first and second iterations of a call-graph generator from static Go code                                             |
|    [`x/codegraph_v3`](./codegraph_v3)    |                                             |                                                  third iteration of a call-graph generator from static Go code                                                   |
|            [`x/conv`](./conv)            |                                             |                                                           standard binary converters for several types                                                           |
|            [`x/cron`](./cron)            |                                             |                                                              cron-like execution scheduler library                                                               |
|            [`x/crop`](./crop)            |                                             |                              an image cropping utility, to divide an input image into X and Y number of tiles (as multiple images)                               |
|         [`x/discord`](./discord)         |                                             |                                                         Discord-focused libraries (webhooks, bots, etc.)                                                         |
|             [`x/dns`](./dns)             |   [🚀](https://github.com/zalgonoise/dns)   |                                           a small but modular DNS server to freely create new routes in your network.                                            |
|        [`x/encoding`](./encoding)        |                                             |                                                             binary encoder libraries (like protobuf)                                                             |
|          [`x/errors`](./errors)          |                                             |                                                  early port of `errors.Join` before it was generally available                                                   |
|            [`x/errs`](./errs)            |                                             |                                     a library to standardize error creation, by defining a Domain, Kind and Entity in a type                                     |
|         [`x/fractal`](./fractal)         |                                             |                                                            Mandelbrot and Julia fractals experiments                                                             |
|             [`x/fts`](./fts)             |   [🚀](https://github.com/zalgonoise/fts)   |                                                SQLite-based full-text search (in-memory and persisted to a file)                                                 |
|            [`x/gbuf`](./gbuf)            |  [🚀](https://github.com/zalgonoise/gbuf)   |                                                 generic buffers like `bytes.Buffer`, but for any specified type.                                                 |
|           [`x/ghttp`](./ghttp)           |                                             |                                 generic HTTP handlers and mux compatible with the standard library, with a simplified structure                                  |
|           [`x/graph`](./graph)           |                                             |                                                                  generic graph data structures                                                                   |
|            [`x/grid`](./grid)            |                                             |                                           a Coordinates, Vectors, 2D Grids, Graphs, and Graph Search set of libraries                                            |
|         [`x/gwriter`](./gwriter)         |   [🚀](https://github.com/zalgonoise/gio)   |                                                         generic I/O library, but for any specified type.                                                         |
|              [`x/is`](./is)              |                                             |                                                very simple (and limited) comparison logic to be used in Go tests                                                 |
|             [`x/lex`](./lex)             |   [🚀](https://github.com/zalgonoise/lex)   |                       generic lexer, following the design in Go's `text/template` and `go/token` implementations, for any specified type.                        |
|             [`x/log`](./log)             |  [🚀](https://github.com/zalgonoise/logx)   |                          second iteration on a structured logger in Go, with a hint from the new `x/log/slog` proposal, with generics.                           |
|        [`x/log/attr`](./log/attr)        |  [🚀](https://github.com/zalgonoise/attr)   |                                            a key-value attribute generic data structure, for any type of requirement.                                            |
|          [`x/logbuf`](./logbuf)          |                                             |                                                     experiments with a buffered slog.Handler implementation                                                      |
|         [`x/mapping`](./mapping)         |                                             |                                                          a dynamic fields mapping library for Go types                                                           |
| [`x/monitoring-tmpl`](./monitoring-tmpl) |                                             |                   Template / example project including all 3 instrumentation signals (metrics, logs, traces) and backends for them as services                   |
|           [`x/parse`](./parse)           |  [🚀](https://github.com/zalgonoise/parse)  | generic parser, as a continuation of [`x/lex`](./lex), that builds a generic AST based on defined tokens and processing rules, and optionally output generation. |
|            [`x/pcap`](./pcap)            |       [🤖](https://chat.openai.com/)        |                                                                        packet capture app                                                                        |
|         [`x/pluslog`](./pluslog)         |                                             |                                                            standard-library `slog` package extensions                                                            |
|             [`x/ptr`](./ptr)             |                                             |                                    generic pointer (and unsafe) utilities library, for conversion, copying, casting and more.                                    |
|            [`x/secr`](./secr)            | [🚀](https://github.com/zalgonoise/cloaki)  |                                 a compact yet module secrets / passwords storage solution, with a Domain Driven Design approach.                                 |
|           [`x/slack`](./slack)           |                                             |                                                          Slack-focused libraries (webhooks, bots, etc.)                                                          |
|         [`x/spanner`](./spanner)         | [🚀](https://github.com/zalgonoise/spanner) |          a simple (but flavored) tracer / spanner solution inspired by OpenTelemetry's, used interchangeably alongside `logx` in `ghttp`, for example.           |
|           [`x/steam`](./steam)           |                                             |                                                                  Steam API and CLI (unofficial)                                                                  |


_Reference_

- 🚀: Ported into its own repository, as linked in the emoji 
- 🤖: ChatGPT experiment; contains code generated with ChatGPT
_________


### Special thanks

_A very heartwarming thank you to [JetBrains](https://www.jetbrains.com/) for their support with their [Open Source Support program](https://jb.gg/OpenSourceSupport)!! Building these libraries, apps, solutions with GoLand makes it such a great developer experience, from the debugger to the tests, coverage and profiling_ ❤️

<div style="display: flex; align-items: center; justify-content: center">
    <a href="https://www.jetbrains.com/" title="JetBrains"><img width="120" height="120" title="JetBrains" src="https://resources.jetbrains.com/storage/products/company/brand/logos/jb_beam.png"></a>
    <a href="https://www.jetbrains.com/go" title="GoLand"><img width="120" height="120" title="GoLand" src="https://resources.jetbrains.com/storage/products/company/brand/logos/GoLand_icon.png"></a>
</div>