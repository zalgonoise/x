package actions

import (
	"errors"

	"github.com/zalgonoise/x/graph/errs"
	"github.com/zalgonoise/x/graph/model"
)

func GraphDepth[T model.ID, I model.Num](g model.Graph[T, I]) int {
	counter := 0
	for g.Parent() != nil {
		counter += 1
		g = g.Parent()
	}
	return counter
}

func BreadthFirstSearch[T model.ID, I model.Num](g model.Graph[T, I], fn func(from, to model.Graph[T, I], weight I) bool, targets ...model.Graph[T, I]) (bool, error) {
	for _, node := range targets {
		if node == nil {
			continue
		}

		for linkedNodes := []model.Graph[T, I]{node}; len(linkedNodes) > 0; {
			target := linkedNodes[0]
			linkedNodes = linkedNodes[1:]
			visited := map[T]struct{}{}

			edges, err := g.Edges(target.ID())
			if err != nil {
				return false, err
			}

			for _, edge := range edges {
				if _, ok := visited[edge.ID()]; ok {
					continue
				}

				w, err := g.Weight(target.ID(), edge.ID())
				if err != nil {
					return false, err
				}
				ok := fn(target, edge, w)
				if !ok {
					return ok, nil
				}
				visited[target.ID()] = struct{}{}
				linkedNodes = append(linkedNodes, edge)
			}
		}
	}
	return true, nil
}

func DepthFirstSearch[T model.ID, I model.Num](g model.Graph[T, I], fn func(from, to model.Graph[T, I], weight I) bool, targets ...model.Graph[T, I]) (bool, error) {
	for _, node := range targets {
		if node == nil {
			continue
		}

		for linkedNodes := []model.Graph[T, I]{node}; len(linkedNodes) > 0; {
			target := linkedNodes[len(linkedNodes)-1]
			linkedNodes = linkedNodes[:len(linkedNodes)-1]
			visited := map[T]struct{}{}

			edges, err := g.Edges(target.ID())
			if err != nil {
				return false, err
			}

			for _, edge := range edges {
				if _, ok := visited[edge.ID()]; ok {
					continue
				}

				w, err := g.Weight(target.ID(), edge.ID())
				if err != nil {
					return false, err
				}
				ok := fn(target, edge, w)
				if !ok {
					return false, nil
				}
				visited[target.ID()] = struct{}{}
				linkedNodes = append(linkedNodes, edge)
			}
		}
	}
	return true, nil
}

func LeafLookup[T model.ID, I model.Num](from model.Graph[T, I], to T) (parent model.Graph[T, I], err error) {
	var (
		g       model.Graph[T, I] = from
		visited                   = map[model.Graph[T, I]]struct{}{from: {}}
	)
	if g == nil {
		return nil, errs.DoesNotExist
	}
	for g.Parent() != nil {
		g = g.Parent()

		// short circuit out - lookup in this graph
		toNode, err := g.Get(to)
		if err == nil && toNode != nil {
			return g, nil
		}
		// continue until it gets to root
		visited[g] = struct{}{}
	}

	return leafLookup(g, to, visited)

}

func leafLookup[T model.ID, I model.Num](root model.Graph[T, I], to T, visited map[model.Graph[T, I]]struct{}) (parent model.Graph[T, I], err error) {
	nodes, err := root.List()
	if err != nil {
		return nil, err
	}

	if _, ok := visited[root]; !ok {
		visited[root] = struct{}{}
	}

	for _, node := range nodes {
		if _, ok := visited[node]; ok {
			continue
		}

		// short circuit out
		if n, err := node.Get(to); err == nil && n != nil {
			return node, nil
		}

		p, err := leafLookup(node, to, visited)
		if err != nil && !errors.Is(err, errs.DoesNotExist) {
			return nil, err
		}
		if p != nil && err == nil {
			return p, nil
		}
	}

	return nil, errs.DoesNotExist
}

func VerifyCycles[T model.ID, I model.Num](from, to model.Graph[T, I]) func(target, edge model.Graph[T, I], weight I) bool {
	return func(target, edge model.Graph[T, I], weight I) bool {
		return to.ID() != target.ID() // fails verification
	}
}
