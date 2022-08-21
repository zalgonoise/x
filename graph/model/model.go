package model

import (
	"github.com/zalgonoise/x/graph/options"
	"golang.org/x/exp/constraints"
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

// // Graph defines the behavior of a graph data structure, with open
// // CRUD operations towards its nodes, their edges, and its ID
// type Graph[T ID, I Int, V any] interface {
// 	// AddNode takes in any number of nodes and adds it to the graph
// 	Add(...Node[T, I, V]) error
// 	// RemoveNode removes any of input nodes from the graph
// 	Remove(...T) error
// 	// GetNode takes in an input ID to a node, and returns it or an error
// 	Get(T) (Node[T, I, V], error)
// 	// Get returns all nodes in a graph, and an error
// 	List() ([]Node[T, I, V], error)

// 	// AddEdge links `from` to `to`, with a set weight
// 	Connect(from, to T, weight I) error
// 	// RemoveEdge unlinks `target` and `edge` nodes
// 	Disconnect(from, to T) error
// 	// GetEdges takes in a node to return a list of nodes linked to it, and and error
// 	Edges(T) ([]Node[T, I, V], error)
// 	// GetWeight gets the weight value of two nodes, if they are connected
// 	Weight(from, to T) (I, error)

// 	ID() T
// 	Value() V
// }

// // Node defines the basic behavior a graph node must have, being able to return an
// // ID
// type Node[T ID, I Int, V any] interface {
// 	ID() T
// 	Parent() Graph[T, I, V]
// 	Link(Graph[T, I, V]) error
// 	Value() V
// }

// Hub defines a multigraph, a graph that hosts other nested graphs
type Graph[T ID, I Int] interface {
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

	ID() T
	Parent() Graph[T, I]
	Link(Graph[T, I], ...options.Setting) error
	Value() any
	// Map() *map[Graph[T, I]]map[Graph[T, I]]I
}
