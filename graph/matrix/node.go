package matrix

import (
	"fmt"

	"github.com/zalgonoise/x/graph/errs"
	"github.com/zalgonoise/x/graph/model"
)

type mapNode[T model.ID, I model.Int] struct {
	id     T
	parent Graph[T, I]
}

func (n *mapNode[T, I]) ID() T {
	return n.id
}
func (n *mapNode[T, I]) Parent() model.Graph[T, I] {
	return n.parent
}
func (n *mapNode[T, I]) Link(gr model.Graph[T, I]) error {
	if gr == nil {
		n.parent = nil
		return nil
	}

	mapGraph, ok := gr.(Graph[T, I])

	if !ok {
		return fmt.Errorf("not a map graph: %w", errs.InvalidType)
	}

	n.parent = mapGraph
	return nil
}

func NewNode[T model.ID, I model.Int](id T) model.Node[T, I] {
	return &mapNode[T, I]{id: id}
}
