package model

import (
	"github.com/zalgonoise/x/graph/options"
	"golang.org/x/exp/constraints"
)

// ID defines the types that can be set as identifiers of a Graph
//
// IDs must be comparable and unique
type ID interface {
	constraints.Ordered
}

// IDer is any type that implements the ID getter method
type IDer[T ID] interface {
	ID() T
}

// Num defines the types that can be set as weights of an Edge
//
// Nums must be integers, floating-point or complex numbers
type Num interface {
	constraints.Integer | constraints.Float | constraints.Complex
}

// Graph defines the behavior of a graph data structure, with open
// CRUD operations towards its nodes and their edges, and also operations
// against the graph itself -- retrieve its ID, its parent, link it to a parent
// and retrieve its underlying value
type Graph[T ID, I Num] interface {
	// AddNode takes in any number of nodes and adds it to the graph
	Add(...Graph[T, I]) error
	// RemoveNode removes any of input nodes from the graph
	Remove(...T) error
	// GetNode takes in an input ID to a node, and returns it or an error
	Get(T) (Graph[T, I], error)
	// Get returns all nodes in a graph, and an error
	List() ([]Graph[T, I], error)

	// AddEdge links `from` to `to`, with a set weight
	Connect(from, to T, weight I) error
	// RemoveEdge unlinks `target` and `edge` nodes
	Disconnect(from, to T) error
	// GetEdges takes in a node to return a list of nodes linked to it, and and error
	Edges(T) ([]Graph[T, I], error)
	// GetWeight gets the weight value of two nodes, if they are connected
	Weight(from, to T) (I, error)

	// ID returns a unique identifier to the node, which must be comparable
	ID() T
	// Parent returns the parent graph storing this node
	Parent() Graph[T, I]
	// Link connects this graph to another, as a child of the input graph
	Link(Graph[T, I], ...options.Setting) error
	// Value returns the value stored when creating this graph or node
	Value() any
	// Config returns the graph's configuration as an options.Setting object
	Config() options.Setting
}
