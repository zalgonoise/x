package graph

// NOTE: this is WIP and a sandbox library for me to test generic implementations
// in graph data structures. Please take it with a grain of salt.

import (
	"reflect"

	"github.com/zalgonoise/x/graph/list"
	"github.com/zalgonoise/x/graph/matrix"
	"github.com/zalgonoise/x/graph/model"
	"github.com/zalgonoise/x/graph/options"
)

func New[T model.ID, I model.Num](id T, value any, opts ...options.Setting) model.Graph[T, I] {
	config := options.New(opts...)
	if config.IDConstraint != nil && config.IDConstraint != reflect.TypeOf(id) {
		return nil
	}

	switch config.GraphType {
	case options.GraphList:
		return list.New[T, I](id, value, config)
	case options.GraphMatrix:
		return matrix.New[T, I](id, value, config)
	default:
		return matrix.New[T, I](id, value, config)
	}
}

func Make[T model.ID, I model.Num](opts options.Setting, nodes ...model.IDer[T]) []model.Graph[T, I] {
	config := options.New(opts)

	graphList := []model.Graph[T, I]{}

	for _, node := range nodes {
		g := New[T, I](node.ID(), node, config)
		if g != nil {
			graphList = append(graphList, g)
		}
	}

	return graphList
}

func Gen[T model.ID, I model.Num](opts options.Setting, nodes ...T) []model.Graph[T, I] {
	config := options.New(opts)

	graphList := []model.Graph[T, I]{}

	for _, node := range nodes {
		g := New[T, I](node, node, config)
		if g != nil {
			graphList = append(graphList, g)
		}
	}

	return graphList
}
