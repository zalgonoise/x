package list

import (
	"github.com/zalgonoise/x/graph/errs"
	"github.com/zalgonoise/x/graph/model"
	"github.com/zalgonoise/x/graph/options"
)

type Graph[T model.ID, I model.Int, V any] interface {
	model.Graph[T, I, V]
	Map() *map[model.Node[T, I, V]][]model.Node[T, I, V]
}

type listGraph[T model.ID, I model.Int, V any] struct {
	id T
	v  V
	n  map[model.Node[T, I, V]][]model.Node[T, I, V]

	conf *options.GraphConfig
}

func NewGraph[T model.ID, I model.Int, V any](id T, v V, conf *options.GraphConfig) model.Graph[T, I, V] {
	return &listGraph[T, I, V]{
		id: id,
		v:  v,
		n:  map[model.Node[T, I, V]][]model.Node[T, I, V]{},

		conf: conf,
	}
}

func (g *listGraph[T, I, V]) Map() *map[model.Node[T, I, V]][]model.Node[T, I, V] {
	return &g.n
}
func (g *listGraph[T, I, V]) ID() T {
	return g.id
}
func (g *listGraph[T, I, V]) Value() V {
	return g.v
}
func (g *listGraph[T, I, V]) Add(nodes ...model.Node[T, I, V]) error {
	ids := []T{}
	for _, k := range nodes {
		ids = append(ids, k.ID())
	}

	return AddNodesToList[T, I, V](g, nodes...)
}
func (g *listGraph[T, I, V]) Remove(nodes ...T) error {
	if g.conf.Immutable {
		return errs.Immutable
	}
	return RemoveNodesFromList[T, I, V](g, nodes...)
}
func (g *listGraph[T, I, V]) Get(node T) (model.Node[T, I, V], error) {
	return GetNodeFromList[T, I, V](g, node)
}
func (g *listGraph[T, I, V]) List() ([]model.Node[T, I, V], error) {
	return ListNodesFromList[T, I, V](g)
}
func (g *listGraph[T, I, V]) Connect(from, to T, weight I) error {
	return AddEdgeInList[T, I, V](g, from, to, 1, g.conf.IsNonDirectional, g.conf.IsNonCyclical)
}
func (g *listGraph[T, I, V]) Disconnect(from, to T) error {
	if g.conf.Immutable {
		return errs.Immutable
	}
	return AddEdgeInList[T, I, V](g, from, to, 0, g.conf.IsNonDirectional, g.conf.IsNonCyclical)
}
func (g *listGraph[T, I, V]) Edges(node T) ([]model.Node[T, I, V], error) {
	return GetEdgesFromListNode[T, I, V](g, node)
}
func (g *listGraph[T, I, V]) Weight(from, to T) (I, error) {
	return GetWeightFromEdgesInList[T, I, V](g, from, to)
}
