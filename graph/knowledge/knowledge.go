package knowledge

import (
	"fmt"

	"github.com/zalgonoise/x/graph/errs"
	"github.com/zalgonoise/x/graph/model"
	"github.com/zalgonoise/x/graph/options"
)

type Graph[T model.ID, I model.Num] interface {
	model.Graph[T, I]
	adjacency() *map[model.Graph[T, I]]map[I][]model.Graph[T, I]
	EdgesWithProperty(node T, weight I) ([]model.Graph[T, I], error)
	Properties(node T) ([]I, error)
}

type knowledgeGraph[T model.ID, I model.Num] struct {
	locked bool
	id     T
	v      any
	n      map[model.Graph[T, I]]map[I][]model.Graph[T, I]
	parent model.Graph[T, I]

	conf *options.GraphConfig
}

func New[T model.ID, I model.Num](id T, v any, conf options.Setting) model.Graph[T, I] {
	c := options.New(conf)
	if c.GraphType != options.GraphKnowledge {
		c.GraphType = options.GraphKnowledge
	}

	return &knowledgeGraph[T, I]{
		id:     id,
		v:      v,
		n:      map[model.Graph[T, I]]map[I][]model.Graph[T, I]{},
		parent: nil,

		conf: c,
	}
}

func (g *knowledgeGraph[T, I]) adjacency() *map[model.Graph[T, I]]map[I][]model.Graph[T, I] {
	return &g.n
}
func (g *knowledgeGraph[T, I]) ID() T {
	return g.id
}
func (g *knowledgeGraph[T, I]) Value() any {
	return g.v
}
func (g *knowledgeGraph[T, I]) Parent() model.Graph[T, I] {
	return g.parent
}
func (g *knowledgeGraph[T, I]) Link(parent model.Graph[T, I], conf ...options.Setting) error {
	if parent == nil {
		g.parent = nil
		return nil
	}

	if parent.ID() == g.ID() {
		return fmt.Errorf("cannot set graph's parent to self: %w", errs.InvalidOperation)
	}
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
func (g *knowledgeGraph[T, I]) Config() options.Setting {
	conf := options.New(g.conf)
	return conf
}
func (g *knowledgeGraph[T, I]) Add(nodes ...model.Graph[T, I]) error {
	if g.locked {
		return errs.ReadOnly
	}
	return AddNodesToList[T, I](g, g.conf, nodes...)
}
func (g *knowledgeGraph[T, I]) Remove(nodes ...T) error {
	if g.locked {
		return errs.ReadOnly
	}
	if g.conf.Immutable {
		return errs.Immutable
	}
	return RemoveNodesFromList[T, I](g, nodes...)
}
func (g *knowledgeGraph[T, I]) Get(node T) (model.Graph[T, I], error) {
	return GetNodeFromList[T, I](g, node)
}
func (g *knowledgeGraph[T, I]) List() ([]model.Graph[T, I], error) {
	return ListNodesFromList[T, I](g)
}
func (g *knowledgeGraph[T, I]) Connect(from, to T, weight I) error {
	if g.locked {
		return errs.ReadOnly
	}
	return AddEdgeInList[T, I](g, from, to, weight, g.conf.IsNonDirectional, g.conf.IsNonCyclical)
}
func (g *knowledgeGraph[T, I]) Disconnect(from, to T) error {
	if g.locked {
		return errs.ReadOnly
	}
	if g.conf.Immutable {
		return errs.Immutable
	}
	return AddEdgeInList[T, I](g, from, to, 0, g.conf.IsNonDirectional, g.conf.IsNonCyclical)
}
func (g *knowledgeGraph[T, I]) Edges(node T) ([]model.Graph[T, I], error) {
	return GetEdgesFromListNode[T, I](g, node)
}
func (g *knowledgeGraph[T, I]) EdgesWithProperty(node T, weight I) ([]model.Graph[T, I], error) {
	return GetEdgesWithProperty[T, I](g, node, weight)
}
func (g *knowledgeGraph[T, I]) Properties(node T) ([]I, error) {
	return GetNodeProperties[T, I](g, node)
}
func (g *knowledgeGraph[T, I]) Weight(from, to T) (I, error) {
	return GetWeightFromEdgesInList[T, I](g, from, to)
}
