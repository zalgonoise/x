package model

import "golang.org/x/exp/constraints"

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
type Graph[T ID, I Int, V any] interface {
	// AddNode takes in any number of nodes and adds it to the graph
	AddNode(...Node[T, I, V]) error
	// RemoveNode removes any of input nodes from the graph
	RemoveNode(...T) error
	// GetNode takes in an input ID to a node, and returns it or an error
	GetNode(T) (Node[T, I, V], error)
	// Get returns all nodes in a graph, and an error
	Get() ([]Node[T, I, V], error)

	// AddEdge links `from` to `to`, with a set weight
	AddEdge(from, to T, weight I) error
	// RemoveEdge unlinks `target` and `edge` nodes
	RemoveEdge(from, to T) error
	// GetEdges takes in a node to return a list of nodes linked to it, and and error
	GetEdges(T) ([]Node[T, I, V], error)
	// GetWeight gets the weight value of two nodes, if they are connected
	GetWeight(from, to T) (I, error)

	ID() T
	Value() V
}

// Node defines the basic behavior a graph node must have, being able to return an
// ID
type Node[T ID, I Int, V any] interface {
	ID() T
	Parent() Graph[T, I, V]
	Link(Graph[T, I, V]) error
	Value() V
}

// Hub defines a multigraph, a graph that hosts other nested graphs
type Hub[T ID, I Int, V any] interface {
	// AddNode takes in any number of nodes and adds it to the graph
	AddNode(...Hub[T, I, V]) error
	// RemoveNode removes any of input nodes from the graph
	RemoveNode(...T) error
	// GetNode takes in an input ID to a node, and returns it or an error
	GetNode(T) (Hub[T, I, V], error)
	// Get returns all nodes in a graph, and an error
	Get() ([]Hub[T, I, V], error)

	// AddEdge links `from` to `to`, with a set weight
	AddEdge(from, to T, weight I) error
	// RemoveEdge unlinks `target` and `edge` nodes
	RemoveEdge(from, to T) error
	// GetEdges takes in a node to return a list of nodes linked to it, and and error
	GetEdges(T) ([]Hub[T, I, V], error)
	// GetWeight gets the weight value of two nodes, if they are connected
	GetWeight(from, to T) (I, error)

	ID() T
	Parent() Hub[T, I, V]
	Link(Hub[T, I, V]) error
	Value() V
	Map() *map[Hub[T, I, V]]map[Hub[T, I, V]]I
}
