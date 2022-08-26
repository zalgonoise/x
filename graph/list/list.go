package list

import (
	"github.com/zalgonoise/x/graph/errs"
	"github.com/zalgonoise/x/graph/model"
	"github.com/zalgonoise/x/graph/options"
)

type Graph[T model.ID, I model.Num] interface {
	model.Graph[T, I]
	adjacency() *map[model.Graph[T, I]][]model.Graph[T, I]
}

type listEdge[T model.ID, I model.Num] struct {
	model.Graph[T, I]
	weight I
}

type listGraph[T model.ID, I model.Num] struct {
	locked bool
	id     T
	v      any
	n      map[model.Graph[T, I]][]model.Graph[T, I]
	parent model.Graph[T, I]

	conf *options.GraphConfig
}

func New[T model.ID, I model.Num](id T, v any, conf options.Setting) model.Graph[T, I] {
	c := &options.GraphConfig{}
	conf.Apply(c)

	return &listGraph[T, I]{
		id:     id,
		v:      v,
		n:      map[model.Graph[T, I]][]model.Graph[T, I]{},
		parent: nil,

		conf: c,
	}
}

func (g *listGraph[T, I]) adjacency() *map[model.Graph[T, I]][]model.Graph[T, I] {
	return &g.n
}
func (g *listGraph[T, I]) ID() T {
	return g.id
}
func (g *listGraph[T, I]) Value() any {
	return g.v
}
func (g *listGraph[T, I]) Parent() model.Graph[T, I] {
	return g.parent
}
func (g *listGraph[T, I]) Link(parent model.Graph[T, I], conf ...options.Setting) error {
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
func (g *listGraph[T, I]) Add(nodes ...model.Graph[T, I]) error {
	if g.locked {
		return errs.ReadOnly
	}
	return AddNodesToList[T, I](g, g.conf, nodes...)
}
func (g *listGraph[T, I]) Remove(nodes ...T) error {
	if g.locked {
		return errs.ReadOnly
	}
	if g.conf.Immutable {
		return errs.Immutable
	}
	return RemoveNodesFromList[T, I](g, nodes...)
}
func (g *listGraph[T, I]) Get(node T) (model.Graph[T, I], error) {
	return GetNodeFromList[T, I](g, node)
}
func (g *listGraph[T, I]) List() ([]model.Graph[T, I], error) {
	return ListNodesFromList[T, I](g)
}
func (g *listGraph[T, I]) Connect(from, to T, weight I) error {
	if g.locked {
		return errs.ReadOnly
	}
	return AddEdgeInList[T, I](g, from, to, weight, g.conf.IsNonDirectional, g.conf.IsNonCyclical)
}
func (g *listGraph[T, I]) Disconnect(from, to T) error {
	if g.locked {
		return errs.ReadOnly
	}
	if g.conf.Immutable {
		return errs.Immutable
	}
	return AddEdgeInList[T, I](g, from, to, 0, g.conf.IsNonDirectional, g.conf.IsNonCyclical)
}
func (g *listGraph[T, I]) Edges(node T) ([]model.Graph[T, I], error) {
	return GetEdgesFromListNode[T, I](g, node)
}
func (g *listGraph[T, I]) Weight(from, to T) (I, error) {
	return GetWeightFromEdgesInList[T, I](g, from, to)
}
