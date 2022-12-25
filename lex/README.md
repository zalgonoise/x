# lex

*a generic lexer written in Go*

_______________

## Overview

`lex` is a lexer for Go, based on the concept of the [`text/template`](https://pkg.go.dev/text/template) lexer, as a generic implementation. The logic behind this lexer is mostly based off of [Rob Pike](https://github.com/robpike)'s talk about [Lexical Scanning in Go](https://www.youtube.com/watch?v=HxaD_trXwRE), which is also seen in the standard library (in [`text/template/parse/lex.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.19.4:src/text/template/parse/lex.go)).


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

### Entities

#### Item

An Item is an object holding a token and a set of values (lexemes) corresponding to that token. It is a key-value data structure, where the value-half is a slice of any type -- which could be populated with any number of items.

Items are created as the lexer tokenizes symbols / units from the input data, and returned to the caller on each `Lexer.NextItem()` call. 

```go
// Item represents a set of any type of tokens identified by a comparable type
type Item[T comparable, V any] struct {
	Type  T
	Value []V
}
```

#### StateFn

A StateFn (state-function) describes a recursive function that will consume each unit of the input data, and either take action, emit tokens, or to finally pass along the data analysis action to a different StateFn. These functions will be implemented by the consumer of the library, describing the steps (their) lexer will take when consuming data. Examples further below.

The point to the StateFn is to keep the Lexer in control of the state, and the StateFn controlling the cursor and actions the Lexer needs to take. This will allow the biggest flexibility possible, as different Lexer StateFn will derive different flavors for the lexer. Thinking of it as a Markdown lexer, the StateFn would define which and how Markdown tokens are classified, emitted, or even ignored.


```go
// StateFn is a recursive function that updates the Lexer state according to a specific
// input token along the lexing / parsing action.
//
// It returns another StateFn or nil, as it consumes each token with a certain logic applied,
// passing along the lexing / parsing to the next StateFn
type StateFn[C comparable, T any, I Item[C, T]] func(l Lexer[C, T, I]) StateFn[C, T, I]
```

#### Lexer

The Lexer is a state-machine that keeps track of the generated Items as the input data is consumed and labled with tokens. It's part cursor, part controller/verifier (within the `StateFn`s), but its main job is keep the state-functions running as the items are consumed, returning lexical items to the caller as they are generated.

The Lexer should be accompanied with a Parser, that consumes the tokenized Items to build a parse tree.

This Lexer exposes methods that should be perceived as utilities for the caller when building `StateFn`s. In reality, when actually *running* the Lexer, the caller will loop through its `NextItem()` method until it hits an EOF token.

The methods in the Lexer will be covered individually below, as well as the design decisions when writing it this way.

```go
// Lexer describes the behavior of a lexer / parser with cursor capabilities
//
// Once spawned, it will consume all tokens as its `NextItem()` method is called,
// returning processed `Item[C, T]` as it goes
//
// Its `Emit()` method pushes items into the stack to be returned, and its `Accept()`,
// `Check()` and `AcceptRun()` methods act as verifiers for a (set of) token(s)
type Lexer[C comparable, T any, I Item[C, T]] interface {

	// Cursor navigates through a slice in a controlled manner, allowing the
	// caller to move forward, backwards, and jump around the slice as they need
	cur.Cursor[T]

	// NextItem processes the tokens sequentially, through the corresponding StateFn
	//
	// As each item is processed, it is returned to the Lexer by `Emit()`, and
	// finally returned to the caller.
	//
	// Note that multiple calls to `NextItem()` should be made when tokenizing input data;
	// usually in a for-loop while the output item is not EOF.
	NextItem() I

	// Emit pushes the set of units identified by token `itemType` to the items channel,
	// that returns it in the NextItem() method.
	//
	// The emitted item will be a subsection of the input data slice, from the lexer's
	// starting index to the current position index.
	//
	// It also sets the lexer's starting index to the current position index.
	Emit(itemType C)

	// Ignore will set the starting point as the current position, ignoring any preceeding units
	Ignore()

	// Backup will rewind the index for the width of the current item
	Backup()

	// Width returns the size of the set of units ready to be emitted with a token
	Width() int

	// Start returns the current starting-point index for when an item is emitted
	Start() int

	// Check passes the current token through the input `verifFn` function as a validator, returning
	// its result
	Check(verifFn func(item T) bool) bool

	// Accept passes the current token through the input `verifFn` function as a validator, returning
	// its result
	//
	// If the validation passes, the cursor has moved one step forward (the unit was consumed)
	//
	// If the validation fails, the cursor rolls back one step
	Accept(verifFn func(item T) bool) bool

	// AcceptRun iterates through all following tokens, passing them through the input `verifFn`
	// function as a validator
	//
	// Once it fails the verification, the cursor is rolledback once, leaving the caller at the unit
	// that failed the verifFn
	AcceptRun(verifFn func(item T) bool)
}
```

## Usage

## Implementing
