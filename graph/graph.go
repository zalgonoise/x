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

func New[T model.ID, I model.Int](id T, value any, opts ...options.Setting) model.Graph[T, I] {
	config, err := options.New(opts...)
	if err != nil {
		return nil
	}
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
