package lex

import (
	cur "github.com/zalgonoise/cur"
)

type Lexer[C comparable, T any, I Item[C, T]] interface {
	cur.Cursor[T]
	NextItem() I
	Ignore()
	Backup()
	Emit(itemType C)
	Accept(verifFn func(item T) bool) bool
	AcceptRun(verifFn func(item T) bool)
}

type lexer[C comparable, T any, I Item[C, T]] struct {
	name  string
	input []T
	start int
	pos   int
	width int
	state StateFn[C, T, I]
	items chan I
}

func NewLexer[C comparable, T any, I Item[C, T]](
	name string,
	initFn StateFn[C, T, I],
	input []T,
) Lexer[C, T, I] {
	l := &lexer[C, T, I]{
		name:  name,
		input: input,
		state: initFn,
		items: make(chan I, 2),
	}
	return l
}

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
			return I{}
		}
	}
}

func (l *lexer[C, T, I]) Ignore() {
	l.start = l.pos
}

func (l *lexer[C, T, I]) Backup() {
	l.pos -= l.width
}

func (l *lexer[C, T, I]) Emit(itemType C) {
	l.items <- I{
		Typ: itemType,
		Val: l.input[l.start:l.pos],
	}
	l.start = l.pos
}

func (l *lexer[C, T, I]) Accept(verifFn func(item T) bool) bool {
	if ok := verifFn(l.Next()); ok {
		return true
	}
	l.Prev()
	return false
}

func (l *lexer[C, T, I]) AcceptRun(verifFn func(item T) bool) {
	for ok := verifFn(l.Next()); ok; {
	}
	l.Backup()
}

func (l *lexer[C, T, I]) Cur() T {
	return l.input[l.pos]
}

func (l *lexer[C, T, I]) Pos() int {
	return l.pos
}

func (l *lexer[C, T, I]) Len() int {
	return len(l.input)
}

func (l *lexer[C, T, I]) Next() T {
	if l.pos >= len(l.input) {
		l.width = 0
		var zero T
		return zero
	}
	l.pos++
	l.width++
	return l.input[l.pos]
}

func (l *lexer[C, T, I]) Prev() T {
	l.pos--
	l.width--
	return l.input[l.pos]
}

func (l *lexer[C, T, I]) Peek() T {
	next := l.Next()
	l.Prev()
	return next
}

func (l *lexer[C, T, I]) Head() T {
	l.pos = 0
	l.start = 0
	return l.input[l.pos]
}

func (l *lexer[C, T, I]) Tail() T {
	l.pos = len(l.input) - 1
	l.start = len(l.input) - 1
	return l.input[l.pos]
}

func (l *lexer[C, T, I]) Idx(idx int) T {
	if idx < 0 {
		return l.Head()
	}
	if idx >= len(l.input) {
		return l.Tail()
	}
	l.pos = idx
	if idx < l.start {
		l.start = idx
	}
	return l.input[l.pos]
}

func (l *lexer[C, T, I]) Offset(amount int) T {
	if l.pos+amount < 0 {
		return l.Head()
	}
	if l.pos+amount >= len(l.input) {
		return l.Tail()
	}
	l.pos += amount
	if l.start-l.pos < 0 {
		l.start += amount
	}
	return l.input[l.pos]
}

func (l *lexer[C, T, I]) PeekIdx(idx int) T {
	if idx >= len(l.input) {
		return l.Tail()
	}
	if idx < 0 {
		return l.Head()
	}
	return l.input[idx]
}

func (l *lexer[C, T, I]) PeekOffset(amount int) T {
	if l.pos+amount >= len(l.input) || l.pos+amount < 0 {
		l.width = 0
		var zero T
		return zero
	}
	return l.input[l.pos+amount]
}

func (l *lexer[C, T, I]) Extract(start, end int) []T {
	if start < 0 {
		start = 0
	}
	if end > len(l.input) {
		end = len(l.input)
	}
	return l.input[start:end]
}
