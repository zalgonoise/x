package join

import (
	"errors"
	"fmt"

	"github.com/zalgonoise/x/graph"
	"github.com/zalgonoise/x/graph/errs"
	"github.com/zalgonoise/x/graph/model"
)

type FilterFunc[T model.ID, I model.Num] func(from, with model.Graph[T, I], node T, edge []T) bool

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

	var fromEdgeMap = map[T][]T{}
	var withEdgeMap = map[T][]T{}

	for _, fromNode := range nodesFrom {
		if filter(from, with, fromNode.ID(), nil) { //&& doesNotExist(out, fromNode.ID(), nil) {
			err = out.Add(fromNode)
			if err != nil {
				return nil, err
			}
		}

		fromEdges, err := from.Edges(fromNode.ID())
		if err != nil {
			return nil, err
		}

		fromEdgeMap[fromNode.ID()] = getIDs(fromEdges)

	}
	for _, withNode := range nodesWith {
		if filter(from, with, withNode.ID(), nil) { // && doesNotExist(out, withNode.ID(), nil) {
			err = out.Add(withNode)
			if err != nil {
				return nil, err
			}
		}
		withEdges, err := with.Edges(withNode.ID())
		if err != nil {
			return nil, err
		}

		withEdgeMap[withNode.ID()] = getIDs(withEdges)
	}

	for k, v := range fromEdgeMap {
		for _, edge := range v {
			w, err := from.Weight(k, edge)
			if err != nil {
				return nil, err
			}

			if filter(from, with, k, []T{edge}) { //&& doesNotExist(out, k, []T{edge}) {
				out.Connect(k, edge, w)
			}
		}
	}

	for k, v := range withEdgeMap {
		for _, edge := range v {
			w, err := with.Weight(k, edge)
			if err != nil {
				return nil, err
			}

			if filter(from, with, k, []T{edge}) { //&& doesNotExist(out, k, []T{edge}) {
				out.Connect(k, edge, w)
			}
		}
	}

	l, _ := out.List()
	fmt.Println(l[0].ID(), l[1].ID())

	return out, nil
}

func OR[T model.ID, I model.Num](from, with model.Graph[T, I], node T, edge []T) bool {
	return true
}

func AND[T model.ID, I model.Num](from, with model.Graph[T, I], node T, edge []T) bool {
	if edge == nil {
		_, ferr := from.Get(node)
		fromOK := errors.Is(ferr, errs.DoesNotExist)

		_, werr := with.Get(node)
		withOK := errors.Is(werr, errs.DoesNotExist)

		if !fromOK && !withOK {
			return true
		}
		return false
	}
	_, ferr := from.Weight(node, edge[0])
	fromOK := errors.Is(ferr, errs.DoesNotExist)

	_, werr := with.Weight(node, edge[0])
	withOK := errors.Is(werr, errs.DoesNotExist)

	if !fromOK && !withOK {
		return true
	}
	return false
}

func XOR[T model.ID, I model.Num](from, with model.Graph[T, I], node T, edge []T) bool {
	if edge == nil {
		_, ferr := from.Get(node)
		fromOK := errors.Is(ferr, errs.DoesNotExist)

		_, werr := with.Get(node)
		withOK := errors.Is(werr, errs.DoesNotExist)

		if (!fromOK && withOK) || (fromOK && !withOK) {
			return true
		}
		return false
	}
	_, ferr := from.Weight(node, edge[0])
	fromOK := errors.Is(ferr, errs.DoesNotExist)

	_, werr := with.Weight(node, edge[0])
	withOK := errors.Is(werr, errs.DoesNotExist)

	if (!fromOK && withOK) || (fromOK && !withOK) {
		return true
	}
	return false
}

func NOT[T model.ID, I model.Num](from, with model.Graph[T, I], node T, edge []T) bool {
	if edge == nil {
		_, ferr := from.Get(node)
		fromOK := errors.Is(ferr, errs.DoesNotExist)

		_, werr := with.Get(node)
		withOK := errors.Is(werr, errs.DoesNotExist)

		if fromOK && withOK {
			return false
		}
		return true
	}
	_, ferr := from.Weight(node, edge[0])
	fromOK := errors.Is(ferr, errs.DoesNotExist)

	_, werr := with.Weight(node, edge[0])
	withOK := errors.Is(werr, errs.DoesNotExist)

	if fromOK && withOK {
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
