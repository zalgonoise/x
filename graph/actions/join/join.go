package join

import (
	"errors"

	"github.com/zalgonoise/x/graph"
	"github.com/zalgonoise/x/graph/errs"
	"github.com/zalgonoise/x/graph/model"
)

type FilterFunc[T model.ID, I model.Num] func(from, with model.Graph[T, I], node T, edge T) bool

func Join[T model.ID, I model.Num](from, with model.Graph[T, I], filter FilterFunc[T, I]) (model.Graph[T, I], error) {
	out := graph.New[T, I](from.ID(), from.Value(), from.Config())

	nodesFrom, err := from.List()
	if err != nil {
		return nil, err
	}
	nodesWith, err := with.List()
	if err != nil {
		return nil, err
	}

	var fromEdgeMap = map[model.Graph[T, I]][]model.Graph[T, I]{}
	var withEdgeMap = map[model.Graph[T, I]][]model.Graph[T, I]{}

	for _, fromNode := range nodesFrom {
		fromEdges, err := from.Edges(fromNode.ID())
		if err != nil {
			return nil, err
		}

		fromEdgeMap[fromNode] = append(fromEdgeMap[fromNode], fromEdges...)

	}
	for _, withNode := range nodesWith {
		withEdges, err := with.Edges(withNode.ID())
		if err != nil {
			return nil, err
		}

		withEdgeMap[withNode] = append(withEdgeMap[withNode], withEdges...)
	}

	for k, v := range fromEdgeMap {
		for _, edge := range v {
			w, err := from.Weight(k.ID(), edge.ID())
			if err != nil {
				return nil, err
			}

			if filter(from, with, k.ID(), edge.ID()) {
				if _, err := out.Get(k.ID()); errors.Is(err, errs.DoesNotExist) {
					out.Add(k)
				}
				if _, err := out.Get(edge.ID()); errors.Is(err, errs.DoesNotExist) {
					out.Add(edge)
				}

				out.Connect(k.ID(), edge.ID(), w)
			}
		}
	}

	for k, v := range withEdgeMap {
		for _, edge := range v {
			w, err := with.Weight(k.ID(), edge.ID())
			if err != nil {
				return nil, err
			}

			if filter(from, with, k.ID(), edge.ID()) {
				if _, err := out.Get(k.ID()); errors.Is(err, errs.DoesNotExist) {
					out.Add(k)
				}
				if _, err := out.Get(edge.ID()); errors.Is(err, errs.DoesNotExist) {
					out.Add(edge)
				}

				out.Connect(k.ID(), edge.ID(), w)
			}
		}
	}

	return out, nil
}

func OR[T model.ID, I model.Num](from, with model.Graph[T, I], node, edge T) bool {
	return true
}

func AND[T model.ID, I model.Num](from, with model.Graph[T, I], node, edge T) bool {
	fw, ferr := from.Weight(node, edge)
	fromOK := errors.Is(ferr, errs.DoesNotExist)

	ww, werr := with.Weight(node, edge)
	withOK := errors.Is(werr, errs.DoesNotExist)

	if !fromOK && !withOK && fw != 0 && ww != 0 {
		return true
	}
	return false
}

func XOR[T model.ID, I model.Num](from, with model.Graph[T, I], node, edge T) bool {
	_, ferr := from.Weight(node, edge)
	_, werr := with.Weight(node, edge)

	if (ferr == nil && werr != nil) || (ferr != nil && werr == nil) {
		return true
	}
	return false
}

func NOT[T model.ID, I model.Num](from, with model.Graph[T, I], node, edge T) bool {
	fw, ferr := from.Weight(node, edge)
	ww, werr := with.Weight(node, edge)

	if ferr == nil && werr == nil && fw == ww {
		return false
	}
	return true
}

func getIDs[T model.ID, I model.Num](g []model.Graph[T, I]) []T {
	out := []T{}

	for _, graph := range g {
		out = append(out, graph.ID())
	}

	return out
}

func doesNotExist[T model.ID, I model.Num](g model.Graph[T, I], node T, edge []T) bool {
	var err error

	if edge == nil {
		_, err = g.Get(node)
	} else {
		_, err = g.Weight(node, edge[0])
	}
	return errors.Is(err, errs.DoesNotExist)
}
