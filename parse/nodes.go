package parse

import "fmt"

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
type Node[C comparable, T any] struct {
	Parent *Node[C, T]
	Type   C
	Value  []T
	Edges  map[C][]*Node[C, T]

	id int
}

// Node creates a new node with type T `typ` and values V `val`, returning its ID
//
// This action updates the tree's position the the new node's ID, and increments the
// tree's `nextID` reference number
func (t *Tree[C, T]) Node(typ C, val ...T) int {
	n := &Node[C, T]{
		Type:  typ,
		Value: val,
		Edges: map[C][]*Node[C, T]{},
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
func (t *Tree[C, T]) Link(from, to int, link C) error {
	var (
		fromNode *Node[C, T]
		toNode   *Node[C, T]
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

// Store places the current node in the input BackupSlot `slot`, in the parse.Tree
//
// If the current position is invalid, the root node (index zero) will be placed instead;
// if that fails too, an error is returned
func (t *Tree[C, T]) Store(slot BackupSlot) error {
	var n *Node[C, T]

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
func (t *Tree[C, T]) Load(slot BackupSlot) *Node[C, T] {
	id, ok := t.backup[slot]
	if !ok || id < 0 {
		return nil
	}
	return t.get(id)
}

// Jump sets the current position in the tree to the node ID loaded from the BackupSlot `slot`,
// returning an OK boolean and an error in case the node does not exist
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
	return true, nil
}

// Cur returns the node at the current position in the tree
func (t *Tree[C, T]) Cur() *Node[C, T] {
	return t.Nodes[t.pos]
}

// Parent returns the node that is parent to the one at the current position in the tree
func (t *Tree[C, T]) Parent() *Node[C, T] {
	n := t.Nodes[t.pos]
	if n == nil {
		return nil
	}
	return n.Parent
}

// Listt returns the child nodes for the one at the current position in the tree, identified by
// link token T `link`
func (t *Tree[C, T]) List(link C) []*Node[C, T] {
	n := t.Nodes[t.pos]
	if n == nil {
		return nil
	}
	return n.Edges[link]
}
