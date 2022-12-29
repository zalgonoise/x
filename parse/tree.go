package parse

import (
	"errors"

	"github.com/zalgonoise/lex"
)

const (
	maxBackup = 5
)

var (
	// ErrInvalidID is a preset error for invalid node IDs
	ErrInvalidID = errors.New("invalid node ID")
	// ErrInvalidID is a preset error for non-existing nodes
	ErrNotFound = errors.New("node was not found")
	// ErrInvalidID is a preset error for cyclical edges in the graph
	ErrCyclicalEdge = errors.New("cyclical edges are not allowed")
)

// BackupSlot is a (weightless) enum type to create defined containers
// for node IDs, so they can be reused or referenced
type BackupSlot struct{}

var (
	Slot0 BackupSlot = struct{}{}
	Slot1 BackupSlot = struct{}{}
	Slot2 BackupSlot = struct{}{}
	Slot3 BackupSlot = struct{}{}
	Slot4 BackupSlot = struct{}{}
)

// Tree is a generic tree data structure to represent a lexical tree
//
// The Tree will buffer tokens of type T from a Lexer, identified by the same
// type of comparable tokens. Ideally, the parser will use diffent tokens when
// classifying an item in the tree -- however the type should be the same
//
// A Tree exposes methods of accessing the root and the current node, as well as
// the current `ParseFn` analyzing the incoming items
type Tree[C comparable, T any] struct {
	Root   *Node[C, T]
	nodes  []*Node[C, T]
	parent *Node[C, T]

	items   []lex.Item[C, T]
	lex     lex.Lexer[C, T, lex.Item[C, T]]
	peek    int
	pos     int
	nextID  int
	backup  map[BackupSlot]int
	parseFn ParseFn[C, T]
}

// New creates a parse.Tree with the input ParseFn `initParse`, initialized with a root node with type T `typ` and
// values V `values`.
//
// It returns the parse.Tree and a channel of lex.Item to receive tokens from a lexer. e.g.:
//
//	func Run(s string) (string, error) {
//	  l := lex.New(initFn, []rune(s))
//	  p, rcv := parse.New(initParse, mytoken.TreeRoot)
//	  for {
//		   i := l.NextItem()
//	    switch i.Type { (...) } // check for errors or EOF
//	    rcv <- i
//	  }
//	}
//
// ...or, if you're feeling wild:
//
//	func Run(s string) (string, error) {
//	  l := lex.New(initFn, []rune(s))
//	  p, rcv := parse.New(initParse, mytoken.TreeRoot)
//	  for {
//		   rcv <- l.NextItem()
//	  }
//	}
func New[C comparable, T any](
	l lex.Lexer[C, T, lex.Item[C, T]],
	initParse ParseFn[C, T],
	typ C,
	values ...T,
) *Tree[C, T] {
	t := &Tree[C, T]{
		pos:    0,
		nodes:  []*Node[C, T]{},
		parent: nil,

		items:   make([]lex.Item[C, T], maxBackup, maxBackup),
		lex:     l,
		peek:    0,
		backup:  map[BackupSlot]int{},
		nextID:  0,
		parseFn: initParse,
	}
	t.Root = t.Node(0, typ, values...)
	return t
}

// Next returns the next Item
func (t *Tree[C, T]) Next() lex.Item[C, T] {
	if t.peek > 0 {
		t.peek--
	} else {
		t.items[0] = t.lex.NextItem()
	}
	return t.items[t.peek]
}

// Peek returns but does not consume the next Item
func (t *Tree[C, T]) Peek() lex.Item[C, T] {
	if t.peek > 0 {
		return t.items[t.peek-1]
	}
	t.peek = 1

	t.items[0] = t.lex.NextItem()
	return t.items[0]
}

// Backup backs the stream up `n` Items
//
// The zeroth Item is already there. Order must be most recent -> oldest
func (t *Tree[C, T]) Backup(items ...lex.Item[C, T]) {
	for idx, item := range items {
		if idx+1 >= maxBackup {
			break
		}
		t.items[idx+1] = item
	}
	t.peek = len(items)
}

// Parse iterates through the incoming lex Items, by calling its `ParseFn`s, until all tokens
// and actions are completed
func (t *Tree[C, T]) Parse() {
	for t.parseFn != nil {
		t.parseFn = t.parseFn(t)
	}
}

// ParseFn is similar to the Lexer's StateFn, as a recursive function that the Parser
// will keep calling during runtime until it runs out of items received from the Lexer
//
// The ParseFn will return another ParseFn that will keep processing the items; which
// could be done in a number of ways (switch statements, helper functions, etc). When
// `nil` is returned, the parser will stop processing lex items
type ParseFn[C comparable, T any] func(t *Tree[C, T]) ParseFn[C, T]

// ProcessFn is a function that can be executed after parsing all the items, and will
// return a known-good type for the developer to work on. This is a step taken after a
// Tree is built
type ProcessFn[C comparable, T any, R any] func(n *Node[C, T]) R
