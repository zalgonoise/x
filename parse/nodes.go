package parse

import (
	"fmt"

	"github.com/zalgonoise/lex"
)

// Node is a generic tree data structure unit. It is presented as a bidirectional
// tree knowledge graph that starts with a root Node (one without a parent) that
// can have zero-to-many children.
//
// It holds a reference to its parent (so that ParseFns can return to the correct
// point in the tree), the item's (joined) lexemes, and edges (if any)
//
// Edges (child nodes) are defined in a list containing the same lexical order as
// received. This allows safely nesting one or mode nodes without losing context of
// the overall structure of the Nodes in the Tree
type Node[C comparable, T any] struct {
	lex.Item[C, T]
	Parent *Node[C, T]
	Edges  []*Node[C, T]

	id int
}

// Node creates a new node from the input Item `item`, returning the created Node
//
// This action updates the tree's position the the new node's ID, and increments the
// tree's `nextID` reference number
//
// It also automatically sets the parent to the node at the previous position index,
// adding the new node as an edge of it.
//
// Note: since creating a Node nests it under the previous Node, it is the responsibility
// of the caller to move back to the parent if that is the intention
func (t *Tree[C, T]) Node(item lex.Item[C, T]) *Node[C, T] {
	var p *Node[C, T] = nil
	if len(t.nodes) > 0 {
		p = t.nodes[t.pos]
	}
	n := &Node[C, T]{
		Item:   item,
		Parent: p,
		Edges:  []*Node[C, T]{},
		id:     t.nextID,
	}
	t.nodes = append(t.nodes, n)
	t.nodes[t.pos].Edges = append(t.nodes[t.pos].Edges, n)
	t.pos = n.id
	t.nextID++
	return n
}

// Store places the current node in the input BackupSlot `slot`, in the parse.Tree
//
// If the current position is invalid, the root node (index zero) will be placed instead;
// if that fails too, an error is returned
func (t *Tree[C, T]) Store(slot BackupSlot) error {
	var n *Node[C, T]

	if n = t.nodes[t.pos]; n != nil {
		t.backup[slot] = n.id
		return nil
	}
	if n = t.nodes[0]; n != nil {
		t.backup[slot] = n.id
		return nil
	}
	return fmt.Errorf("failed to load node on current position and on position zero: %w", ErrNotFound)
}

// Load returns the node stored in the input BackupSlot `slot`, or nil if either its ID is
// invalid or if the slot is empty
//
// If successful, this action will also clear the BackupSlot `slot`
func (t *Tree[C, T]) Load(slot BackupSlot) *Node[C, T] {
	id, ok := t.backup[slot]
	if !ok || id < 0 {
		return nil
	}
	delete(t.backup, slot)
	return t.get(id)
}

// Jump sets the current position in the tree to the node ID loaded from the BackupSlot `slot`,
// returning an OK boolean and an error in case the node does not exist
//
// If successful, this action will also clear the BackupSlot `slot`
func (t *Tree[C, T]) Jump(slot BackupSlot) (bool, error) {
	id, ok := t.backup[slot]
	if !ok || id < 0 {
		return false, fmt.Errorf("failed to find any nodes in this backup slot: %w", ErrNotFound)
	}
	n := t.get(id)
	if n == nil {
		return false, fmt.Errorf("failed to load node with ID %d: %w", id, ErrNotFound)
	}
	t.pos = id
	delete(t.backup, slot)
	return true, nil
}

// Set places the input node's position as the current one in the Tree
func (t *Tree[C, T]) Set(n *Node[C, T]) error {
	if !t.exists(n) {
		return ErrNotFound
	}
	t.pos = n.id
	return nil
}

// List returns all top-level nodes, under the Tree's Root
func (t *Tree[C, T]) List() []*Node[C, T] {
	return t.Root.Edges
}

// Cur returns the node at the current position in the tree
func (t *Tree[C, T]) Cur() *Node[C, T] {
	if t.pos >= len(t.nodes) {
		return nil
	}
	return t.nodes[t.pos]
}

// Parent returns the node that is parent to the one at the current position in the tree
func (t *Tree[C, T]) Parent() *Node[C, T] {
	if t.pos >= len(t.nodes) {
		return nil
	}
	n := t.nodes[t.pos]
	if n == nil {
		return nil
	}
	return n.Parent
}

// get returns the node with ID `id`, or nil if it does not exist
func (t *Tree[C, T]) get(id int) *Node[C, T] {
	if id < 0 || id >= len(t.nodes) {
		return nil
	}
	return t.nodes[id]
}

func (t *Tree[C, T]) exists(n *Node[C, T]) bool {
	return n != nil && n.id >= 0 && n.id <= len(t.nodes)-1
}
