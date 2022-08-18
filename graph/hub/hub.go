package hub

import (
	"encoding/json"

	"github.com/zalgonoise/x/graph/model"
)

type hubGraph[T model.ID, I model.Int, V any] struct {
	id     T
	v      V
	n      map[model.Hub[T, I, V]]map[model.Hub[T, I, V]]I
	parent model.Hub[T, I, V]

	isNonDirectional bool
	isNonCyclical    bool
}

func New[T model.ID, I model.Int, V any](id T, value V, isNonDir, isNonCyc bool) model.Hub[T, I, V] {
	return &hubGraph[T, I, V]{
		id:     id,
		v:      value,
		n:      map[model.Hub[T, I, V]]map[model.Hub[T, I, V]]I{},
		parent: nil,

		isNonDirectional: isNonDir,
		isNonCyclical:    isNonCyc,
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
	isNonDir := g.isNonDirectional
	isNonCyc := g.isNonCyclical
	return AddEdgeInMap[T, I, V](g, from, to, weight, isNonDir, isNonCyc)
}
func (g *hubGraph[T, I, V]) RemoveEdge(from, to T) error {
	return AddEdgeInMap[T, I, V](g, from, to, 0, g.isNonCyclical, g.isNonCyclical)
}
func (g *hubGraph[T, I, V]) GetEdges(node T) ([]model.Hub[T, I, V], error) {
	return GetEdgesFromMapNode[T, I, V](g, node)
}
func (g *hubGraph[T, I, V]) GetWeight(from, to T) (I, error) {
	return GetWeightFromEdgesInMap[T, I, V](g, from, to)
}

type output[T model.ID, I model.Int, V any] struct {
	ID    T             `json:"id"`
	Data  V             `json:"data,omitempty"`
	Nodes map[T]map[T]I `json:"nodes,omitempty"`
}

func (g *hubGraph[T, I, V]) String() string {
	var out = output[T, I, V]{
		ID:    g.ID(),
		Data:  g.Value(),
		Nodes: map[T]map[T]I{},
	}

	for ko, vo := range g.n {
		innerMap := map[T]I{}
		for ki, vi := range vo {
			innerMap[ki.ID()] = vi
		}
		out.Nodes[ko.ID()] = innerMap
	}

	b, _ := json.Marshal(out)
	return string(b)
}
