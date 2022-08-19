package matrix

import (
	"github.com/zalgonoise/x/graph/errs"
	"github.com/zalgonoise/x/graph/model"
	"github.com/zalgonoise/x/graph/options"
)

type Graph[T model.ID, I model.Int, V any] interface {
	model.Graph[T, I, V]
	Map() *map[model.Node[T, I, V]]map[model.Node[T, I, V]]I
}

type mapGraph[T model.ID, I model.Int, V any] struct {
	id T
	v  V
	n  map[model.Node[T, I, V]]map[model.Node[T, I, V]]I

	conf *options.GraphConfig
}

func NewGraph[T model.ID, I model.Int, V any](id T, v V, conf *options.GraphConfig) model.Graph[T, I, V] {
	return &mapGraph[T, I, V]{
		id: id,
		v:  v,
		n:  map[model.Node[T, I, V]]map[model.Node[T, I, V]]I{},

		conf: conf,
	}
}

func (g *mapGraph[T, I, V]) Map() *map[model.Node[T, I, V]]map[model.Node[T, I, V]]I {
	return &g.n
}
func (g *mapGraph[T, I, V]) ID() T {
	return g.id
}
func (g *mapGraph[T, I, V]) Value() V {
	return g.v
}
func (g *mapGraph[T, I, V]) AddNode(nodes ...model.Node[T, I, V]) error {
	return AddNodesToMap[T, I, V](g, nodes...)
}
func (g *mapGraph[T, I, V]) RemoveNode(nodes ...T) error {
	if g.conf.Immutable {
		return errs.Immutable
	}
	return RemoveNodesFromMap[T, I, V](g, nodes...)
}
func (g *mapGraph[T, I, V]) GetNode(node T) (model.Node[T, I, V], error) {
	return GetNodeFromMap[T, I, V](g, node)
}
func (g *mapGraph[T, I, V]) Get() ([]model.Node[T, I, V], error) {
	return GetKeysFromMap[T, I, V](g)
}
func (g *mapGraph[T, I, V]) AddEdge(from, to T, weight I) error {
	if g.conf.IsUnweighted {
		return AddEdgeInMap[T, I, V](g, from, to, 1, g.conf.IsNonDirectional, g.conf.IsNonCyclical)
	}
	return AddEdgeInMap[T, I, V](g, from, to, weight, g.conf.IsNonDirectional, g.conf.IsNonCyclical)
}
func (g *mapGraph[T, I, V]) RemoveEdge(from, to T) error {
	if g.conf.Immutable {
		return errs.Immutable
	}
	return AddEdgeInMap[T, I, V](g, from, to, 0, g.conf.IsNonDirectional, g.conf.IsNonCyclical)
}
func (g *mapGraph[T, I, V]) GetEdges(node T) ([]model.Node[T, I, V], error) {
	return GetEdgesFromMapNode[T, I, V](g, node)
}
func (g *mapGraph[T, I, V]) GetWeight(from, to T) (I, error) {
	return GetWeightFromEdgesInMap[T, I, V](g, from, to)
}
