package matrix

import (
	"github.com/zalgonoise/x/graph/model"
)

type Graph[T model.ID, I model.Int] interface {
	model.Graph[T, I]
	Map() *map[model.Node[T, I]]map[model.Node[T, I]]I
}

type mapGraph[T model.ID, I model.Int] struct {
	id T
	n  map[model.Node[T, I]]map[model.Node[T, I]]I

	isNonDirectional bool
	isNonCyclical    bool
}

func NewGraph[T model.ID, I model.Int](id T, isNonDir, isNonCyc bool) model.Graph[T, I] {
	return &mapGraph[T, I]{
		id: id,
		n:  map[model.Node[T, I]]map[model.Node[T, I]]I{},

		isNonDirectional: isNonDir,
		isNonCyclical:    isNonCyc,
	}
}

func (g *mapGraph[T, I]) Map() *map[model.Node[T, I]]map[model.Node[T, I]]I {
	return &g.n
}
func (g *mapGraph[T, I]) ID() T {
	return g.id
}
func (g *mapGraph[T, I]) AddNode(nodes ...model.Node[T, I]) error {
	return AddNodesToMap[T, I](g, nodes...)
}
func (g *mapGraph[T, I]) RemoveNode(nodes ...T) error {
	return RemoveNodesFromMap[T, I](g, nodes...)
}
func (g *mapGraph[T, I]) GetNode(node T) (model.Node[T, I], error) {
	return GetNodeFromMap[T, I](g, node)
}
func (g *mapGraph[T, I]) Get() ([]model.Node[T, I], error) {
	return GetKeysFromMap[T, I](g)
}
func (g *mapGraph[T, I]) AddEdge(from, to T, weight I) error {
	isNonDir := g.isNonDirectional
	isNonCyc := g.isNonCyclical
	return AddEdgeInMap[T, I](g, from, to, weight, isNonDir, isNonCyc)
}
func (g *mapGraph[T, I]) RemoveEdge(from, to T) error {
	return AddEdgeInMap[T, I](g, from, to, 0, g.isNonCyclical, g.isNonCyclical)
}
func (g *mapGraph[T, I]) GetEdges(node T) ([]model.Node[T, I], error) {
	return GetEdgesFromMapNode[T, I](g, node)
}
func (g *mapGraph[T, I]) GetWeight(from, to T) (I, error) {
	return GetWeightFromEdgesInMap[T, I](g, from, to)
}
