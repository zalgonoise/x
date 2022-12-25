package lex

import (
	cur "github.com/zalgonoise/cur"
)

// Lexer describes the behavior of a lexer / parser with cursor capabilities
//
// Once spawned, it will consume all tokens as its `NextItem()` method is called,
// returning processed `Item[C, T]` as it goes
//
// Its `Emit()` method pushes items into the stack to be returned, and its `Accept()`
// and `AcceptRun()` methods act as verifiers for a (set of) token(s)
type Lexer[C comparable, T any, I Item[C, T]] interface {
	// Cursor navigates through a slice in a controlled manner, allowing the
	// caller to move forward, backwards, and jump around the slice as they need
	cur.Cursor[T]
	// NextItem processes the tokens sequentially, through the corresponding StateFn
	//
	// As each item is processed, it is returned to the Lexer by `NextItem()`, which is
	// finally returned to the caller when finished
	NextItem() I
	// Ignore will scroll the starting point to the current position
	Ignore()
	// Backup will rewind the index for the width of the current item
	Backup()
	// Width returns the size for the set of tokens
	Width() int
	// Start returns the current starting-point index for when an item is emitted
	Start() int
	// Emit pushes the set of tokens from the starting point to the current position to the
	// receiver, identified by C `itemType`
	Emit(itemType C)
	// Check passes the current token through the input `verifFn` function as a validator, returning
	// its result
	Check(verifFn func(item T) bool) bool
	// Accept passes the next token through the input `verifFn` function as a validator, returning
	// its result while also rewinding to the current position
	Accept(verifFn func(item T) bool) bool
	// AcceptRun iterates through all following tokens, passing them through the input `verifFn`
	// function as a validator, returning its result while also rewinding to the current position
	AcceptRun(verifFn func(item T) bool)
}

type lexer[C comparable, T any, I Item[C, T]] struct {
	input []T
	start int
	pos   int
	state StateFn[C, T, I]
	items chan I
}

// New creates a new lexer with the base / starting StateFn and input data
func New[C comparable, T any, I Item[C, T]](
	initFn StateFn[C, T, I],
	input []T,
) Lexer[C, T, I] {
	if len(input) == 0 {
		return nil
	}

	l := &lexer[C, T, I]{
		input: input,
		state: initFn,
		items: make(chan I, 2),
	}
	return l
}

// NextItem processes the tokens sequentially, through the corresponding StateFn
//
// As each item is processed, it is returned to the Lexer by `NextItem()`, which is
// finally returned to the caller when finished
func (l *lexer[C, T, I]) NextItem() I {
	for {
		select {
		case item := <-l.items:
			return item
		default:
			if l.state != nil {
				l.state = l.state(l)
				continue
			}
			close(l.items)
			var eof I
			return eof
		}
	}
}

// Ignore will scroll the starting point to the current position
func (l *lexer[C, T, I]) Ignore() {
	l.start = l.pos
}

// Backup will rewind the index for the width of the current item
func (l *lexer[C, T, I]) Backup() {
	l.pos = l.start
}

// Width returns the size for the set of tokens
func (l *lexer[C, T, I]) Width() int {
	return l.pos - l.start
}

// Emit pushes the set of tokens from the starting point to the current position to the
// receiver, identified by C `itemType`
func (l *lexer[C, T, I]) Emit(itemType C) {
	l.items <- I{
		Type:  itemType,
		Value: l.input[l.start:l.pos],
	}
	l.start = l.pos
}

// Accept passes the next token through the input `verifFn` function as a validator, returning
// its result while also rewinding to the current position
func (l *lexer[C, T, I]) Accept(verifFn func(item T) bool) bool {
	if ok := verifFn(l.Next()); ok {
		return true
	}
	l.Prev()
	return false
}

// Check passes the current token through the input `verifFn` function as a validator, returning
// its result
func (l *lexer[C, T, I]) Check(verifFn func(item T) bool) bool {
	return verifFn(l.Cur())
}

// AcceptRun iterates through all following tokens, passing them through the input `verifFn`
// function as a validator, returning its result while also rewinding to the current position
func (l *lexer[C, T, I]) AcceptRun(verifFn func(item T) bool) {
	for verifFn(l.Next()) {
	}
	l.Prev()
	return
}

// Cur returns the same indexed item in the slice
func (l *lexer[C, T, I]) Cur() T {
	if l.pos >= len(l.input) {
		var eof T
		return eof
	}
	return l.input[l.pos]
}

// Pos returns the current position in the cursor
func (l *lexer[C, T, I]) Pos() int {
	return l.pos
}

// Len returns the total size of the underlying slice
func (l *lexer[C, T, I]) Len() int {
	return len(l.input)
}

// Start returns the current starting-point index for when an item is emitted
func (l *lexer[C, T, I]) Start() int {
	return l.start
}

// Next returns the next item in the slice, or the last index if already
// at the tail
func (l *lexer[C, T, I]) Next() T {
	if l.pos >= len(l.input) {
		var eof T
		return eof
	}
	l.pos++
	return l.input[l.pos-1]
}

// Prev returns the previous item in the slice, or the first index if
// already at the head
func (l *lexer[C, T, I]) Prev() T {
	if l.pos-1 < 0 {
		var eof T
		return eof
	}
	l.pos--
	return l.input[l.pos]
}

// Peek returns the next indexed item without advancing the cursor
func (l *lexer[C, T, I]) Peek() T {
	if l.pos+1 >= len(l.input) {
		var eof T
		return eof
	}
	return l.input[l.pos+1]
}

// Head returns to the beginning of the slice
func (l *lexer[C, T, I]) Head() T {
	l.pos = 0
	l.start = 0
	return l.input[l.pos]
}

// Tail jumps to the end of the slice
func (l *lexer[C, T, I]) Tail() T {
	l.pos = len(l.input) - 1
	l.start = len(l.input) - 1
	return l.input[l.pos]
}

// Idx jumps to the specific index `idx` in the slice
//
// If the input index is below 0, the head of the slice is returned;
// If the input index is greater than the size of the slice, the tail of the slice
// is returned
func (l *lexer[C, T, I]) Idx(idx int) T {
	if idx < 0 {
		var eof T
		return eof
	}
	if idx >= len(l.input) {
		var eof T
		return eof
	}
	l.pos = idx
	if idx < l.start {
		l.start = idx
	}
	return l.input[l.pos]
}

// Offset advances or rewinds `amount` steps in the slice, be it a positive or negative
// input.
//
// If the result offset is below 0, the head of the slice is returned;
// If the result offset is greater than the size of the slice, the tail of the slice
// is returned
func (l *lexer[C, T, I]) Offset(amount int) T {
	if l.pos+amount < 0 {
		var eof T
		return eof
	}
	if l.pos+amount >= len(l.input) {
		var eof T
		return eof
	}
	l.pos += amount
	if l.pos-l.start < 0 {
		l.start += amount
	}
	return l.input[l.pos]
}

// PeekIdx returns the next indexed item without advancing the cursor,
// with the index `idx`
func (l *lexer[C, T, I]) PeekIdx(idx int) T {
	if idx >= len(l.input) {
		var eof T
		return eof
	}
	if idx < 0 {
		var eof T
		return eof
	}
	return l.input[idx]
}

// PeekOffset returns the next indexed item without advancing the cursor,
// with offset `amount`
func (l *lexer[C, T, I]) PeekOffset(amount int) T {
	if l.pos+amount >= len(l.input) || l.pos+amount < 0 {
		var eof T
		return eof
	}
	return l.input[l.pos+amount]
}

// Extract returns a slice from index `start` to index `end`
func (l *lexer[C, T, I]) Extract(start, end int) []T {
	if start < 0 {
		start = 0
	}
	if end > len(l.input) {
		end = len(l.input)
	}
	return l.input[start:end]
}
