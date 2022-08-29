package graph

import (
	"github.com/zalgonoise/x/graph/errs"
	"github.com/zalgonoise/x/graph/list"
	"github.com/zalgonoise/x/graph/matrix"
	"github.com/zalgonoise/x/graph/model"
	"github.com/zalgonoise/x/graph/options"
)

func From[T model.ID, I model.Num](g model.Graph[T, I]) (model.Graph[T, I], error) {
	conf := options.New(g.Config())

	// graph type inverter
	switch g.(type) {
	case matrix.Graph[T, I]:
		conf.GraphType = options.GraphList
	case list.Graph[T, I]:
		conf.GraphType = options.GraphMatrix
	default:
		return nil, errs.InvalidType
	}

	// initialize output graph with new config
	out := New[T, I](g.ID(), g.Value(), conf)

	// get all nodes, add them to output graph
	nodes, err := g.List()
	if err != nil {
		return nil, err
	}
	out.Add(nodes...)

	// range through all nodes to get edges and their weights
	// connect those nodes with said weight
	for _, n := range nodes {
		edges, err := g.Edges(n.ID())
		if err != nil {
			return nil, err
		}

		for _, e := range edges {
			w, err := g.Weight(n.ID(), e.ID())
			if err != nil {
				return nil, err
			}

			out.Connect(n.ID(), e.ID(), w)
		}
	}

	return out, nil
}
