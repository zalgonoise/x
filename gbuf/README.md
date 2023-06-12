# gbuf

*A generic buffer library for Go*

_________


## Overview

Similar to [`zalgonoise/gio`](https://github.com/zalgonoise/gio), this library extends the usefulness of the interfaces and functions exposed by the standard library (for byte buffers) with generics. The core functionality of a reader and buffer (something that reads) should be common amongst any type, provided that the implementation can handle the calls (to read, to write, or whatever action).

This way, despite not having the same fluidity as the standard library's implementation in some levels (there is no type-specific method such as `WriteString` and `WriteByte`), it promises to allow the same API to be transposed to other defined types.

Other than this, all functionality from the standard library is present in this generic I/O library.

### Why generics?

Generics are great for when there is a solid algorithm that serves for many types, and can be abstracted enough to work without major workarounds; and this approach to a buffers library is very idiomatic and so simple (the Go way). Of course, the standard library's implementation has some other packages in mind that work together with `bytes`, namely UTF-encoding and properly handling runes. The approach with generics will limit the potential that shines in the original implementation, one way or the other (simply with the fact that if you need handle different types, you need to convert them yourself).

But all in all, it was a great exercise to practice using generics. Maybe I will just use this library once or twice, maybe it will be actually useful for some. I am just in it for the ride. :)


## Disclaimer

This library will mirror all logic from Go's (standard) `bytes` and `container` libraries; and change the `[]byte` implementation with a generic `T any` and `[]T` implementation. There are no changes in the actual logic in the library.
________________

### Added Features

Besides recently adding `container` library generic implementations (heap, list and ring); I've also extended the concept of the circular buffer with the `RingBuffer` type, that is a circular buffer with an `io.Reader` / `io.Writer` implementation (and goodies). Another type similar to this one is `RingFilter` which allows passing a slice of the unread items on each cycle to a given `func([]T) error` -- that allows filtering data / chaining readers / building data processing pipelines.