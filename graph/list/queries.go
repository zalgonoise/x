package list

import (
	"github.com/zalgonoise/x/graph/model"
)

func BreadthFirstSearch[T model.ID, I model.Int, V any](g Graph[T, I, V], fn func(from, to model.Node[T, I, V], weight I) bool, targets ...model.Node[T, I, V]) (bool, error) {
	for _, node := range targets {
		if node == nil {
			continue
		}

		for linkedNodes := []model.Node[T, I, V]{node}; len(linkedNodes) > 0; {
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

func DepthFirstSearch[T model.ID, I model.Int, V any](g Graph[T, I, V], fn func(from, to model.Node[T, I, V], weight I) bool, targets ...model.Node[T, I, V]) (bool, error) {
	for _, node := range targets {
		if node == nil {
			continue
		}

		for linkedNodes := []model.Node[T, I, V]{node}; len(linkedNodes) > 0; {
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

func VerifyCycles[T model.ID, I model.Int, V any](from, to model.Node[T, I, V]) func(target, edge model.Node[T, I, V], weight I) bool {
	return func(target, edge model.Node[T, I, V], weight I) bool {
		return to.ID() != target.ID() // fails verification
	}
}
