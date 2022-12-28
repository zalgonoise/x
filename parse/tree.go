package parse

import (
	"errors"
	"fmt"

	"github.com/zalgonoise/cur"
	"github.com/zalgonoise/lex"
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
type Tree[T comparable, V any] struct {
	Nodes  map[int]*Node[T, V]
	Cursor cur.Cursor[lex.Item[T, V]]

	items   *[]lex.Item[T, V]
	pos     int
	nextID  int
	recv    chan lex.Item[T, V]
	backup  map[BackupSlot]int
	parseFn ParseFn[T, V]
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
func New[T comparable, V any](initParse ParseFn[T, V], typ T, values ...V) (*Tree[T, V], chan lex.Item[T, V]) {
	items := &[]lex.Item[T, V]{}
	t := &Tree[T, V]{
		pos:    0,
		Nodes:  map[int]*Node[T, V]{},
		Cursor: cur.Ptr(items),

		items:   items,
		recv:    make(chan lex.Item[T, V]),
		backup:  map[BackupSlot]int{},
		nextID:  0,
		parseFn: initParse,
	}
	_ = t.Node(typ, values...)

	go t.run()
	return t, t.recv
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
	Edges  map[T][]*Node[T, V]

	id int
}

// Store places the current node in the input BackupSlot `slot`, in the parse.Tree
//
// If the current position is invalid, the root node (index zero) will be placed instead;
// if that fails too, an error is returned
func (t *Tree[T, V]) Store(slot BackupSlot) error {
	var n *Node[T, V]

	if n = t.Nodes[t.pos]; n != nil {
		t.backup[slot] = n.id
		return nil
	}
	if n = t.Nodes[0]; n != nil {
		t.backup[slot] = n.id
		return nil
	}
	return fmt.Errorf("failed to load node on current position and on position zero: %w", ErrNotFound)
}

// Load returns the node stored in the input BackupSlot `slot`, or nil if either its ID is
// invalid or if the slot is empty
func (t *Tree[T, V]) Load(slot BackupSlot) *Node[T, V] {
	id, ok := t.backup[slot]
	if !ok || id < 0 {
		return nil
	}
	return t.get(id)
}

// Jump sets the current position in the tree to the node ID loaded from the BackupSlot `slot`,
// returning an OK boolean and an error in case the node does not exist
func (t *Tree[T, V]) Jump(slot BackupSlot) (bool, error) {
	id, ok := t.backup[slot]
	if !ok || id < 0 {
		return false, fmt.Errorf("failed to find any nodes in this backup slot: %w", ErrNotFound)
	}
	n := t.get(id)
	if n == nil {
		return false, fmt.Errorf("failed to load node with ID %d: %w", id, ErrNotFound)
	}
	t.pos = id
	return true, nil
}

// Cur returns the node at the current position in the tree
func (t *Tree[T, V]) Cur() *Node[T, V] {
	return t.Nodes[t.pos]
}

// Parent returns the node that is parent to the one at the current position in the tree
func (t *Tree[T, V]) Parent() *Node[T, V] {
	n := t.Nodes[t.pos]
	if n == nil {
		return nil
	}
	return n.Parent
}

// Listt returns the child nodes for the one at the current position in the tree, identified by
// link token T `link`
func (t *Tree[T, V]) List(link T) []*Node[T, V] {
	n := t.Nodes[t.pos]
	if n == nil {
		return nil
	}
	return n.Edges[link]
}

// Node creates a new node with type T `typ` and values V `val`, returning its ID
//
// This action updates the tree's position the the new node's ID, and increments the
// tree's `nextID` reference number
func (t *Tree[T, V]) Node(typ T, val ...V) int {
	n := &Node[T, V]{
		Type:  typ,
		Value: val,
		Edges: map[T][]*Node[T, V]{},
		id:    t.nextID,
	}
	t.Nodes[n.id] = n
	t.pos = n.id
	t.nextID++
	return n.id
}

// Link creates an edge between two nodes, identified by link token T `link`
//
// It returns an error if either node does not exist; or if the action tries to introduce
// a cyclical edge in the graph
func (t *Tree[T, V]) Link(from, to int, link T) error {
	var (
		fromNode *Node[T, V]
		toNode   *Node[T, V]
	)
	if fromNode = t.get(from); fromNode == nil {
		return fmt.Errorf("failed to get from-node: %w", ErrNotFound)
	}
	if toNode = t.get(to); toNode == nil {
		return fmt.Errorf("failed to get to-node: %w", ErrNotFound)
	}
	for p := fromNode.Parent; p != nil; {
		if p.id == to {
			return ErrCyclicalEdge
		}
	}

	fromEdges := fromNode.Edges[link]
	for _, e := range fromEdges {
		if e.id == to {
			// no action required; already is an edge
			return nil
		}
	}
	fromNode.Edges[link] = append(fromNode.Edges[link], toNode)
	toNode.Parent = fromNode
	return nil
}

// run keeps receiving lex.Items and stores them in the tree (as the lexer runs), while at the same time
// keeps processing the stored items via the configured ParseFn
func (t *Tree[T, V]) run() {
	var eof T
	for {
		select {
		case i := <-t.recv:
			*t.items = append(*t.items, i)
			if i.Type == eof {
				close(t.recv)
			}
		default:
			if t.parseFn != nil {
				t.parseFn = t.parseFn(t)
			}
			return
		}
	}
}

// get returns the node with ID `id`, or nil if it does not exist
func (t *Tree[T, V]) get(id int) *Node[T, V] {
	if n, ok := t.Nodes[id]; ok {
		return n
	}
	return nil
}

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
