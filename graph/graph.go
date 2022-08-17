package graph

// NOTE: this is WIP and a sandbox library for me to test generic implementations
// in graph data structures. Please take it with a grain of salt.

import (
	"github.com/zalgonoise/x/graph/matrix"
	"github.com/zalgonoise/x/graph/model"
	"github.com/zalgonoise/x/graph/options"
)

func New[T model.ID, I model.Int](id T, opt options.Code) model.Graph[T, I] {
	isList, isNonDir, isNonCyc := options.Parse(opt)
	if !isList {
		// build adjacency matrix
		return matrix.NewGraph[T, I](id, isNonDir, isNonCyc)
	} else {
		// build adjacency list
		return nil
	}
}
