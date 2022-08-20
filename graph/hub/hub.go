package hub

import (
	"github.com/zalgonoise/x/graph/errs"
	"github.com/zalgonoise/x/graph/model"
	"github.com/zalgonoise/x/graph/options"
)

type hubGraph[T model.ID, I model.Int] struct {
	locked bool
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
func (g *hubGraph[T, I]) Link(parent model.Hub[T, I], conf ...options.Setting) error {
	g.parent = parent
	if g.conf.ReadOnly {
		g.locked = true
	}

	n := len(conf)
	if n == 1 {
		conf[0].Apply(g.conf)
	} else if n > 1 {
		options.MultiOption(conf...).Apply(g.conf)
	}

	return nil
}
func (g *hubGraph[T, I]) Value() any {
	return g.v
}
func (g *hubGraph[T, I]) Map() *map[model.Hub[T, I]]map[model.Hub[T, I]]I {
	return &g.n
}
func (g *hubGraph[T, I]) Add(nodes ...model.Hub[T, I]) error {
	if g.locked {
		return errs.ReadOnly
	}
	return AddNodesToMap[T, I](g, g.conf, nodes...)
}
func (g *hubGraph[T, I]) Remove(nodes ...T) error {
	if g.locked {
		return errs.ReadOnly
	}
	if g.conf.Immutable {
		return errs.Immutable
	}
	return RemoveNodesFromMap[T, I](g, nodes...)
}
func (g *hubGraph[T, I]) Get(node T) (model.Hub[T, I], error) {
	return GetNodeFromMap[T, I](g, node)
}
func (g *hubGraph[T, I]) List() ([]model.Hub[T, I], error) {
	return ListNodesFromMap[T, I](g)
}
func (g *hubGraph[T, I]) Connect(from, to T, weight I) error {
	if g.locked {
		return errs.ReadOnly
	}
	if g.conf.IsUnweighted {
		return AddEdgeInMap[T, I](g, from, to, 1, g.conf.IsNonDirectional, g.conf.IsNonCyclical)
	}
	return AddEdgeInMap[T, I](g, from, to, weight, g.conf.IsNonDirectional, g.conf.IsNonCyclical)
}
func (g *hubGraph[T, I]) Disconnect(from, to T) error {
	if g.locked {
		return errs.ReadOnly
	}
	if g.conf.Immutable {
		return errs.Immutable
	}
	return AddEdgeInMap[T, I](g, from, to, 0, g.conf.IsNonDirectional, g.conf.IsNonCyclical)
}
func (g *hubGraph[T, I]) Edges(node T) ([]model.Hub[T, I], error) {
	return GetEdgesFromMapNode[T, I](g, node)
}
func (g *hubGraph[T, I]) Weight(from, to T) (I, error) {
	return GetWeightFromEdgesInMap[T, I](g, from, to)
}
