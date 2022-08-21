package linkedlist

import (
	"github.com/zalgonoise/x/graph/errs"
	"github.com/zalgonoise/x/graph/model"
	"github.com/zalgonoise/x/graph/options"
)

type Graph[T model.ID, I model.Num] interface {
	model.Graph[T, I]
	nextGraph(model.Graph[T, I]) model.Graph[T, I]
	parentGraph(model.Graph[T, I]) model.Graph[T, I]
}

type linkedList[T model.ID, I model.Num] struct {
	locked bool
	id     T
	v      any
	next   model.Graph[T, I]
	parent model.Graph[T, I]

	conf *options.GraphConfig
}

func New[T model.ID, I model.Num](id T, v any, conf options.Setting) model.Graph[T, I] {
	c := &options.GraphConfig{}
	conf.Apply(c)

	list := &linkedList[T, I]{
		id:     id,
		v:      v,
		next:   nil,
		parent: nil,

		conf: c,
	}

	return list
}

func (g *linkedList[T, I]) nextGraph(in model.Graph[T, I]) model.Graph[T, I] {
	if in != nil {
		g.next = in
		return nil
	}
	return g.next
}

func (g *linkedList[T, I]) parentGraph(in model.Graph[T, I]) model.Graph[T, I] {
	if in != nil {
		g.parent = in
		return nil
	}
	return g.parent
}
func (g *linkedList[T, I]) ID() T {
	return g.id
}
func (g *linkedList[T, I]) Value() any {
	return g.v
}
func (g *linkedList[T, I]) Parent() model.Graph[T, I] {
	return g.parent
}
func (g *linkedList[T, I]) Link(parent model.Graph[T, I], conf ...options.Setting) error {
	g.parentGraph(parent)

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
func (g *linkedList[T, I]) Add(nodes ...model.Graph[T, I]) error {
	if g.locked {
		return errs.ReadOnly
	}
	return AddNodesToList[T, I](g, g.conf, nodes...)
}
func (g *linkedList[T, I]) Remove(nodes ...T) error {
	if g.locked {
		return errs.ReadOnly
	}
	if g.conf.Immutable {
		return errs.Immutable
	}
	return RemoveNodesFromList[T, I](g, nodes...)
}
func (g *linkedList[T, I]) Get(node T) (model.Graph[T, I], error) {
	return GetNodeFromList[T, I](g, node)
}
func (g *linkedList[T, I]) List() ([]model.Graph[T, I], error) {
	return ListNodesFromList[T, I](g)
}
func (g *linkedList[T, I]) Connect(from, to T, weight I) error {
	return errs.InvalidOperation
}
func (g *linkedList[T, I]) Disconnect(from, to T) error {
	return errs.InvalidOperation
}
func (g *linkedList[T, I]) Edges(node T) ([]model.Graph[T, I], error) {
	return nil, errs.InvalidOperation
}
func (g *linkedList[T, I]) Weight(from, to T) (I, error) {
	return 0, errs.InvalidOperation
}
