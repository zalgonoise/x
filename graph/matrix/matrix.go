package matrix

import (
	"github.com/zalgonoise/x/graph/errs"
	"github.com/zalgonoise/x/graph/model"
	"github.com/zalgonoise/x/graph/options"
)

type Graph[T model.ID, I model.Num] interface {
	model.Graph[T, I]
	adjancy() *map[model.Graph[T, I]]map[model.Graph[T, I]]I
}

type mapGraph[T model.ID, I model.Num] struct {
	locked bool
	id     T
	v      any
	n      map[model.Graph[T, I]]map[model.Graph[T, I]]I
	parent model.Graph[T, I]

	conf *options.GraphConfig
}

func New[T model.ID, I model.Num](id T, v any, conf options.Setting) model.Graph[T, I] {
	c := &options.GraphConfig{}
	conf.Apply(c)

	return &mapGraph[T, I]{
		id:     id,
		v:      v,
		n:      map[model.Graph[T, I]]map[model.Graph[T, I]]I{},
		parent: nil,

		conf: c,
	}
}

func (g *mapGraph[T, I]) adjancy() *map[model.Graph[T, I]]map[model.Graph[T, I]]I {
	return &g.n
}
func (g *mapGraph[T, I]) ID() T {
	return g.id
}
func (g *mapGraph[T, I]) Value() any {
	return g.v
}
func (g *mapGraph[T, I]) Parent() model.Graph[T, I] {
	return g.parent
}
func (g *mapGraph[T, I]) Link(parent model.Graph[T, I], conf ...options.Setting) error {
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
func (g *mapGraph[T, I]) Add(nodes ...model.Graph[T, I]) error {
	if g.locked {
		return errs.ReadOnly
	}
	return AddNodesToMap[T, I](g, g.conf, nodes...)
}
func (g *mapGraph[T, I]) Remove(nodes ...T) error {
	if g.locked {
		return errs.ReadOnly
	}
	if g.conf.Immutable {
		return errs.Immutable
	}
	return RemoveNodesFromMap[T, I](g, nodes...)
}
func (g *mapGraph[T, I]) Get(node T) (model.Graph[T, I], error) {
	return GetNodeFromMap[T, I](g, node)
}
func (g *mapGraph[T, I]) List() ([]model.Graph[T, I], error) {
	return ListNodesFromMap[T, I](g)
}
func (g *mapGraph[T, I]) Connect(from, to T, weight I) error {
	if g.locked {
		return errs.ReadOnly
	}
	if g.conf.IsUnweighted {
		return AddEdgeInMap[T, I](g, from, to, 1, g.conf.IsNonDirectional, g.conf.IsNonCyclical)
	}
	return AddEdgeInMap[T, I](g, from, to, weight, g.conf.IsNonDirectional, g.conf.IsNonCyclical)
}
func (g *mapGraph[T, I]) Disconnect(from, to T) error {
	if g.locked {
		return errs.ReadOnly
	}
	if g.conf.Immutable {
		return errs.Immutable
	}
	return AddEdgeInMap[T, I](g, from, to, 0, g.conf.IsNonDirectional, g.conf.IsNonCyclical)
}
func (g *mapGraph[T, I]) Edges(node T) ([]model.Graph[T, I], error) {
	return GetEdgesFromMapNode[T, I](g, node)
}
func (g *mapGraph[T, I]) Weight(from, to T) (I, error) {
	return GetWeightFromEdgesInMap[T, I](g, from, to)
}
