package parse

import "github.com/zalgonoise/x/lex"

// Tree is a generic tree data structure to represent a lexical tree
//
// The Tree will buffer tokens of type T from a Lexer, identified by the same
// type of comparable tokens. Ideally, the parser will use diffent tokens when
// classifying an item in the tree -- however the type should be the same
//
// A Tree exposes methods of accessing the root and the current node, as well as
// the current `ParseFn` analyzing the incoming items
type Tree[T comparable, V any] struct {
	Root  *Node[T, V]
	Pos   *Node[T, V]
	Items []lex.Item[T, V]

	parseFn ParseFn[T, V]
	nextID  uint
}

// Node is a generic tree data structure unit. It is presented as a bidirectional
// tree knowledge graph that starts with a root Node (one without a parent) that
// can have zero-to-many children.
//
// It holds a reference to its parent (so that ParseFns can return to the correct
// point in the tree), the item's (joined) lexemes, and children nodes (if any)
//
// Children nodes are defined in a map identified by comparable token T that holds
// a list of zero-to-many Nodes. This allows many approaches (such as using T for
// a weight or an index), but mainly to serve as a relationship indicator
type Node[T comparable, V any] struct {
	Parent *Node[T, V]
	Type   T
	Value  []V
	Nodes  map[T][]*Node[T, V]
	id     uint
}

// LinkFn is a function that joins two nodes together, where Node `a` is parent of
// Node `b`, with the link value `v` of type T.
//
// It is a defined function type that is unimplemented to allow the developer to
// choose how Nodes are linked -- limits; thresholds; checks; etc.
type LinkFn[T comparable, V any] func(a, b *Node[T, V], link T)

// ParseFn is similar to the Lexer's StateFn, as a recursive function that the Parser
// will keep calling during runtime until it runs out of items received from the Lexer
//
// The ParseFn will return another ParseFn that will keep processing the items; which
// could be done in a number of ways (switch statements, helper functions, etc). When
// `nil` is returned, the parser will stop processing lex items
type ParseFn[T comparable, V any] func(t *Tree[T, V]) ParseFn[T, V]

// ProcessFn is a function that can be executed after parsing all the items, and will
// return a known-good type for the developer to work on. This is a step taken after a
// Tree is built
type ProcessFn[T comparable, V any, R any] func(t *Tree[T, V]) R
