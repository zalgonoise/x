package graph

// NOTE: this is WIP and a sandbox library for me to test generic implementations
// in graph data structures. Please take it with a grain of salt.

import (
	"errors"

	"golang.org/x/exp/constraints"
)

var (
	ErrAlreadyExists error = errors.New("already exists")
	ErrNotGraph      error = errors.New("input is not a graph")
	ErrNotNode       error = errors.New("input is not a node")
	ErrDoesNotExist  error = errors.New("target does not exist")
	ErrInvalidType   error = errors.New("invalid input type")
	ErrIDMismatch    error = errors.New("input IDs do not match")
)

// ID defines the types that can be set as identifiers to a Node
//
// IDs must be comparable and unique
type ID interface {
	constraints.Ordered
}

type Int interface {
	constraints.Integer | constraints.Float
}

// Graph defines the behavior of a graph data structure, with open
// CRUD operations towards its nodes, their edges, and its ID
type Graph[T ID, I Int] interface {
	// AddNode takes in any number of nodes and adds it to the graph
	AddNode(...Node[T]) error
	// RemoveNode removes any of input nodes from the graph
	RemoveNode(...T) error
	// GetNode takes in an input ID to a node, and returns it or an error
	GetNode(T) (Node[T], error)
	// Get returns all nodes in a graph, and an error
	Get() ([]Node[T], error)

	// AddEdge links `from` to `to`, with a set weight
	AddEdge(from, to T, weight I) error
	// RemoveEdge unlinks `target` and `edge` nodes
	RemoveEdge(target, edge T) error
	// GetEdges takes in a node to return a list of nodes linked to it, and and error
	GetEdges(T) ([]Node[T], error)

	ID() T
}

// Node defines the basic behavior a graph node must have, being able to return an
// ID
type Node[T ID] interface {
	ID() T
}
