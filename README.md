### `/x/` 

Experimental Go libraries _(take it with a grain of salt)_

______________

### Concept

This repository contains projects, experiments and studies when developing new Go libraries and applications.

Its content ranges from HTTP / net utilities, generics in Go, binary encoding, to audio processing and much more in between. In `/x/` many ideas are born, and some eventually ported to their own repositories.  

This is the place for trying something new, to study it and amplify an initial idea. It's a place to be unbounded and unconstrained; to both succeed and to fail.


_________

### Index 


|             Directory              |                                                                                                   Description                                                                                                    |                                                                                                    
|:----------------------------------:|:----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------:|
|        [`x/audio`](./audio)        |                                                                 digital signal processing tools for audio, including WAV encoding and decoding.                                                                  |
|    [`x/benchmark`](./benchmark)    |                                                             several benchmark tests to compare `x` implementations with third-party implementations                                                              |
|       [`x/bezier`](./bezier)       |                                                                               PNG to SVG converter, as a quick ChatGPT experiment                                                                                |
|           [`x/cb`](./cb)           |                                                                                 experiments on a Circuit Breaker implementation                                                                                  |
|    [`x/codegraph`](./codegraph)    |                                                                    first and second iterations of a call-graph generator from static Go code                                                                     |
| [`x/codegraph_v3`](./codegraph_v3) |                                                                          third iteration of a call-graph generator from static Go code                                                                           |
|         [`x/conv`](./conv)         |                                                                                   standard binary converters for several types                                                                                   |
|         [`x/crop`](./crop)         |                                                      an image cropping utility, to divide an input image into X and Y number of tiles (as multiple images)                                                       |
|          [`x/dns`](./dns)          |                                            a small but modular DNS server to freely create new routes in your network. [_Ported_](https://github.com/zalgonoise/dns)                                             |
|     [`x/encoding`](./encoding)     |                                                                                     binary encoder libraries (like protobuf)                                                                                     |
|       [`x/errors`](./errors)       |                                                                          early port of `errors.Join` before it was generally available                                                                           |
|      [`x/fractal`](./fractal)      |                                                                                    Mandelbrot and Julia fractals experiments                                                                                     |
|         [`x/gbuf`](./gbuf)         |                                                 generic buffers like `bytes.Buffer`, but for any specified type. [_Ported_](https://github.com/zalgonoise/gbuf)                                                  |
|        [`x/ghttp`](./ghttp)        |                                                         generic HTTP handlers and mux compatible with the standard library, with a simplified structure                                                          |
|        [`x/graph`](./graph)        |                                                                                          generic graph data structures                                                                                           |
|      [`x/gwriter`](./gwriter)      |                                                          generic I/O library, but for any specified type. [_Ported_](https://github.com/zalgonoise/gio)                                                          |
|          [`x/lex`](./lex)          |                        generic lexer, following the design in Go's `text/template` and `go/token` implementations, for any specified type. [_Ported_](https://github.com/zalgonoise/lex)                         |
|          [`x/log`](./log)          |                           second iteration on a structured logger in Go, with a hint from the new `x/log/slog` proposal, with generics. [_Ported_](https://github.com/zalgonoise/logx)                           |
|     [`x/log/attr`](./log/attr)     |                                            a key-value attribute generic data structure, for any type of requirement. [_Ported_](https://github.com/zalgonoise/attr)                                             |
|        [`x/parse`](./parse)        | generic parser, as a continuation of [`x/lex`](./lex), that builds a generic AST based on defined tokens and processing rules, and optionally output generation. [_Ported_](https://github.com/zalgonoise/parse) |
|         [`x/pcap`](./pcap)         |                                                                                packet capture app, as a quick ChatGPT experiment                                                                                 |
|          [`x/ptr`](./ptr)          |                                                                  generic pointer utilities library, for conversion, copying, casting and more.                                                                   |
|         [`x/secr`](./secr)         |                                a compact yet module secrets / passwords storage solution, with a Domain Driven Design approach. [_Ported_](https://github.com/zalgonoise/cloaki)                                 |
|      [`x/spanner`](./spanner)      |         a simple (but flavored) tracer / spanner solution inspired by OpenTelemetry's, used interchangeably alongside `logx` in `ghttp`, for example. [_Ported_](https://github.com/zalgonoise/spanner)          |

_________


### Special thanks

_A very heartwarming thank you to [JetBrains](https://www.jetbrains.com/) for their support with their [Open Source Support program](https://jb.gg/OpenSourceSupport)!! Building these libraries, apps, solutions with GoLand makes it such a great developer experience, from the debugger to the tests, coverage and profiling_ ❤️

<div style="display: flex; align-items: center; justify-content: center">
    <a href="https://www.jetbrains.com/" title="JetBrains"><img width="120" height="120" title="JetBrains" src="https://resources.jetbrains.com/storage/products/company/brand/logos/jb_beam.png"></a>
    <a href="https://www.jetbrains.com/go" title="GoLand"><img width="120" height="120" title="GoLand" src="https://resources.jetbrains.com/storage/products/company/brand/logos/GoLand_icon.png"></a>
</div>