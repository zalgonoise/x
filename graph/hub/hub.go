package hub

import (
	"github.com/zalgonoise/x/graph/model"
	"github.com/zalgonoise/x/graph/options"
)

type hubGraph[T model.ID, I model.Int, V any] struct {
	id     T
	v      V
	n      map[model.Hub[T, I, V]]map[model.Hub[T, I, V]]I
	parent model.Hub[T, I, V]

	conf *options.GraphConfig
}

func New[T model.ID, I model.Int, V any](id T, value V, config *options.GraphConfig) model.Hub[T, I, V] {
	return &hubGraph[T, I, V]{
		id:     id,
		v:      value,
		n:      map[model.Hub[T, I, V]]map[model.Hub[T, I, V]]I{},
		parent: nil,

		conf: config,
	}
}

func (g *hubGraph[T, I, V]) ID() T {
	return g.id
}
func (g *hubGraph[T, I, V]) Parent() model.Hub[T, I, V] {
	return g.parent
}
func (g *hubGraph[T, I, V]) Link(parent model.Hub[T, I, V]) error {
	g.parent = parent
	return nil
}
func (g *hubGraph[T, I, V]) Value() V {
	return g.v
}
func (g *hubGraph[T, I, V]) Map() *map[model.Hub[T, I, V]]map[model.Hub[T, I, V]]I {
	return &g.n
}
func (g *hubGraph[T, I, V]) AddNode(nodes ...model.Hub[T, I, V]) error {
	return AddNodesToMap[T, I, V](g, nodes...)
}
func (g *hubGraph[T, I, V]) RemoveNode(nodes ...T) error {
	return RemoveNodesFromMap[T, I, V](g, nodes...)
}
func (g *hubGraph[T, I, V]) GetNode(node T) (model.Hub[T, I, V], error) {
	return GetNodeFromMap[T, I, V](g, node)
}
func (g *hubGraph[T, I, V]) Get() ([]model.Hub[T, I, V], error) {
	return GetKeysFromMap[T, I, V](g)
}
func (g *hubGraph[T, I, V]) AddEdge(from, to T, weight I) error {
	return AddEdgeInMap[T, I, V](g, from, to, weight, g.conf.IsNonDirectional, g.conf.IsNonCyclical)
}
func (g *hubGraph[T, I, V]) RemoveEdge(from, to T) error {
	return AddEdgeInMap[T, I, V](g, from, to, 0, g.conf.IsNonDirectional, g.conf.IsNonCyclical)
}
func (g *hubGraph[T, I, V]) GetEdges(node T) ([]model.Hub[T, I, V], error) {
	return GetEdgesFromMapNode[T, I, V](g, node)
}
func (g *hubGraph[T, I, V]) GetWeight(from, to T) (I, error) {
	return GetWeightFromEdgesInMap[T, I, V](g, from, to)
}
