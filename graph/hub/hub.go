package hub

import (
	"github.com/zalgonoise/x/graph/model"
	"github.com/zalgonoise/x/graph/options"
)

type hubGraph[T model.ID, I model.Int] struct {
	id     T
	v      any
	n      map[model.Hub[T, I]]map[model.Hub[T, I]]I
	parent model.Hub[T, I]

	conf *options.GraphConfig
}

func New[T model.ID, I model.Int](id T, value any, config *options.GraphConfig) model.Hub[T, I] {
	return &hubGraph[T, I]{
		id:     id,
		v:      value,
		n:      map[model.Hub[T, I]]map[model.Hub[T, I]]I{},
		parent: nil,

		conf: config,
	}
}

func (g *hubGraph[T, I]) ID() T {
	return g.id
}
func (g *hubGraph[T, I]) Parent() model.Hub[T, I] {
	return g.parent
}
func (g *hubGraph[T, I]) Link(parent model.Hub[T, I]) error {
	g.parent = parent
	return nil
}
func (g *hubGraph[T, I]) Value() any {
	return g.v
}
func (g *hubGraph[T, I]) Map() *map[model.Hub[T, I]]map[model.Hub[T, I]]I {
	return &g.n
}
func (g *hubGraph[T, I]) AddNode(nodes ...model.Hub[T, I]) error {
	return AddNodesToMap[T, I](g, nodes...)
}
func (g *hubGraph[T, I]) RemoveNode(nodes ...T) error {
	return RemoveNodesFromMap[T, I](g, nodes...)
}
func (g *hubGraph[T, I]) GetNode(node T) (model.Hub[T, I], error) {
	return GetNodeFromMap[T, I](g, node)
}
func (g *hubGraph[T, I]) Get() ([]model.Hub[T, I], error) {
	return GetKeysFromMap[T, I](g)
}
func (g *hubGraph[T, I]) AddEdge(from, to T, weight I) error {
	return AddEdgeInMap[T, I](g, from, to, weight, g.conf.IsNonDirectional, g.conf.IsNonCyclical)
}
func (g *hubGraph[T, I]) RemoveEdge(from, to T) error {
	return AddEdgeInMap[T, I](g, from, to, 0, g.conf.IsNonDirectional, g.conf.IsNonCyclical)
}
func (g *hubGraph[T, I]) GetEdges(node T) ([]model.Hub[T, I], error) {
	return GetEdgesFromMapNode[T, I](g, node)
}
func (g *hubGraph[T, I]) GetWeight(from, to T) (I, error) {
	return GetWeightFromEdgesInMap[T, I](g, from, to)
}
