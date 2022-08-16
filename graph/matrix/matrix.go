package matrix

import (
	"github.com/zalgonoise/x/graph/model"
)

type Graph[T model.ID, I model.Int] interface {
	model.Graph[T, I]
	Map() *map[model.Node[T]]map[model.Node[T]]I
	Keys() *map[T]model.Node[T]
}

type mapNode[T model.ID] struct {
	id T
}

func (n *mapNode[T]) ID() T {
	return n.id
}

func NewNode[T model.ID](id T) model.Node[T] {
	return &mapNode[T]{id: id}
}

type mapGraph[T model.ID, I model.Int] struct {
	id   T
	n    map[model.Node[T]]map[model.Node[T]]I
	keys map[T]model.Node[T]

	addNodeFn func(g Graph[T, I], nodes ...model.Node[T]) error
	addEdgeFn func(g Graph[T, I], from, to T, weight I) error
}

func NewGraph[T model.ID, I model.Int](id T, dirOpt, cycOpt bool) model.Graph[T, I] {
	// set defaults
	var (
		addNodeFn func(g Graph[T, I], nodes ...model.Node[T]) error = AddNodesToMap[T, I]
		addEdgeFn func(g Graph[T, I], from, to T, weight I) error   = AddEdgeInMapUni[T, I]
	)

	// non-directional
	if dirOpt {
		addEdgeFn = AddEdgeInMapBi[T, I]
	}

	// non-cyclical
	if cycOpt {
		// non-implemented
		// addNodeFn = nil
	}

	return &mapGraph[T, I]{
		id:   id,
		n:    map[model.Node[T]]map[model.Node[T]]I{},
		keys: map[T]model.Node[T]{},

		// defaults
		addNodeFn: addNodeFn,
		addEdgeFn: addEdgeFn,
	}
}

func (g *mapGraph[T, I]) Map() *map[model.Node[T]]map[model.Node[T]]I {
	return &g.n
}
func (g *mapGraph[T, I]) Keys() *map[T]model.Node[T] {
	return &g.keys
}
func (g *mapGraph[T, I]) ID() T {
	return g.id
}
func (g *mapGraph[T, I]) AddNode(nodes ...model.Node[T]) error {
	return g.addNodeFn(g, nodes...)
}
func (g *mapGraph[T, I]) RemoveNode(nodes ...T) error {
	return RemoveNodesFromMap[T, I](g, nodes...)
}
func (g *mapGraph[T, I]) GetNode(node T) (model.Node[T], error) {
	return GetNodeFromMap[T, I](g, node)
}
func (g *mapGraph[T, I]) Get() ([]model.Node[T], error) {
	return GetKeysFromMap[T, I](g)
}
func (g *mapGraph[T, I]) AddEdge(from, to T, weight I) error {
	return g.addEdgeFn(g, from, to, weight)
}
func (g *mapGraph[T, I]) RemoveEdge(target, edge T) error {
	return AddEdgeInMapUni[T, I](g, target, edge, 0)
}
func (g *mapGraph[T, I]) GetEdges(node T) ([]model.Node[T], error) {
	return GetEdgesFromMapNode[T, I](g, node)
}
