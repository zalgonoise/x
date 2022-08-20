package graph

// NOTE: this is WIP and a sandbox library for me to test generic implementations
// in graph data structures. Please take it with a grain of salt.

import (
	"reflect"

	"github.com/zalgonoise/x/graph/hub"
	"github.com/zalgonoise/x/graph/list"
	"github.com/zalgonoise/x/graph/matrix"
	"github.com/zalgonoise/x/graph/model"
	"github.com/zalgonoise/x/graph/options"
)

func NewGraph[T model.ID, I model.Int, V any](id T, value V, opts ...options.Setting) model.Graph[T, I, V] {
	config, err := options.New(opts...)
	if err != nil {
		return nil
	}

	if config.IDConstraint != nil && config.IDConstraint != reflect.TypeOf(id) {
		return nil
	}

	switch config.GraphType {
	case options.GraphList:
		return list.NewGraph[T, I](id, value, config)
	case options.GraphMatrix:
		return matrix.NewGraph[T, I](id, value, config)
	case options.GraphHub, options.GraphNode:
		return nil
	default:
		return matrix.NewGraph[T, I](id, value, config)
	}
}

func NewHub[T model.ID, I model.Int](id T, value any, opts ...options.Setting) model.Hub[T, I] {
	config, err := options.New(opts...)
	if err != nil {
		panic(err)
		// return nil
	}

	if config.GraphType != options.GraphHub && config.GraphType != options.UnsetType {
		return nil
	}
	return hub.New[T, I](id, value, config)
}

func NewNode[T model.ID, I model.Int, V any](id T, value V, opts ...options.Setting) model.Node[T, I, V] {
	config, err := options.New(opts...)
	if err != nil {
		return nil
	}

	if config.GraphType != options.GraphNode && config.GraphType != options.UnsetType {
		return nil
	}
	return matrix.NewNode[T, I](id, value)
}
