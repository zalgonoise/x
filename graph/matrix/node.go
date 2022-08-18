package matrix

import (
	"fmt"

	"github.com/zalgonoise/x/graph/errs"
	"github.com/zalgonoise/x/graph/model"
)

type mapNode[T model.ID, I model.Int, V any] struct {
	id     T
	v      V
	parent Graph[T, I, V]
}

func (n *mapNode[T, I, V]) ID() T {
	return n.id
}
func (n *mapNode[T, I, V]) Parent() model.Graph[T, I, V] {
	return n.parent
}
func (n *mapNode[T, I, V]) Value() V {
	return n.v
}
func (n *mapNode[T, I, V]) Link(gr model.Graph[T, I, V]) error {
	if gr == nil {
		n.parent = nil
		return nil
	}

	mapGraph, ok := gr.(Graph[T, I, V])

	if !ok {
		return fmt.Errorf("not a map graph: %w", errs.InvalidType)
	}

	n.parent = mapGraph
	return nil
}

func NewNode[T model.ID, I model.Int, V any](id T, value V) model.Node[T, I, V] {
	return &mapNode[T, I, V]{id: id, v: value}
}
