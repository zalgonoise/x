# lex

*a generic lexer written in Go*

_______________

## Overview

`lex` is a lexer for Go, based on the concept of the [`text/template`](https://pkg.go.dev/text/template) lexer, as a generic implementation. The logic behind this lexer is mostly based off of Rob Pike's talk about [Lexical Scanning in Go](https://www.youtube.com/watch?v=HxaD_trXwRE), which is also seen in the standard library (in [`text/template/parse/lex.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.19.4:src/text/template/parse/lex.go)).


The idea behind implementing a generic algorithm for a lexer came from trying to build a graph (data structure) representing the logic blocks in a Go file. Watching the talk above was a breath of fresh air when it came to the design of the lexer and its simple approach. So, it would be nice to leverage this algorithm for the Go code graph idea from before. By making the logic generic, one could implement an `Item` type to hold a defined token type, and a set of (any type of) values and `StateFn` state-functions to tokenize input data. In concept this works for any given type, as the point is to label elements of a slice with identifying tokens, that will be processed into a parse tree (with a specific parser implementation).

Caveats are precisely using very *open* types for this implementation. The `text/template` lexer will, for example, define its EOF token as `-1` -- a constant found in the [`lex.go` file](https://cs.opensource.google/go/go/+/refs/tags/go1.19.4:src/text/template/parse/lex.go;l=93). For this implementation, the lexer will return a zero-value token, so the caller should prepare their token types considering that the zero value will be reserved for EOF. Scrolling through the input will use the `pos int` position, and will not have a width -- because the lexer will consume the input as a list of the defined data type. concerns about the width need to be handled in the types or `StateFn` implementation, not the Lexer.

Additionally, as it is exposed as an interface, it introduces a few helper methods to either validate input data, navigate through the index, and controlling the cursor of the slice. It is implementing the [`cur.Cursor[T any] interface`](https://github.com/zalgonoise/cur).

## Installation 

> Note: this library is not ready out-of-the box! You will need to implement your own `StateFn` state-functions with defined types. This repo will expose simple examples to understand the flow of the lexer, below.

You can add this library to your Go project with a `go get` command:

```
go get github.com/zalgonoise/lex
```

## Features

## Usage

## Implementing
