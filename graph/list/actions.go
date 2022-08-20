package list

import (
	"encoding/json"
	"fmt"

	"github.com/zalgonoise/x/graph/errs"
	"github.com/zalgonoise/x/graph/model"
)

func getKeysFromList[T model.ID, I model.Int, V any](g Graph[T, I, V]) map[T]model.Node[T, I, V] {
	m := *g.Map()
	keyMap := map[T]model.Node[T, I, V]{}

	for k := range m {
		keyMap[k.ID()] = k
	}
	return keyMap
}

func AddNodesToList[T model.ID, I model.Int, V any](g Graph[T, I, V], nodes ...model.Node[T, I, V]) error {
	m := g.Map()
	n := *m

	for _, node := range nodes {

		if _, ok := n[node]; ok {
			return errs.AlreadyExists
		}

		n[node] = []model.Node[T, I, V]{}

		// link node to graph
		node.Link(g)
	}

	m = &n
	return nil
}

func RemoveNodesFromList[T model.ID, I model.Int, V any](g Graph[T, I, V], ids ...T) error {
	m := g.Map()
	n := *m

	for _, id := range ids {
		node, err := g.Get(id)
		if err != nil {
			return err
		}

		if _, ok := n[node]; !ok {
			return err
		}

		// disconnect any edges
		for innerNode, innerEdges := range n {
			if innerNode == node {
				continue
			}
			for _, e := range innerEdges {
				if e == node {
					err := g.Disconnect(innerNode.ID(), node.ID())
					if err != nil {
						return err
					}
				}
			}
		}

		// remove node
		delete(n, node)

		// unlink node from graph
		node.Link(nil)
	}

	m = &n
	return nil
}

func GetNodeFromList[T model.ID, I model.Int, V any](g Graph[T, I, V], node T) (model.Node[T, I, V], error) {
	k := getKeysFromList(g)

	n, ok := k[node]
	if !ok {
		return nil, errs.DoesNotExist
	}

	return n, nil
}

func ListNodesFromList[T model.ID, I model.Int, V any](g Graph[T, I, V]) ([]model.Node[T, I, V], error) {
	m := *g.Map()

	out := []model.Node[T, I, V]{}

	for k := range m {
		out = append(out, k)
	}

	return out, nil
}

func AddEdgeInList[T model.ID, I model.Int, V any](g Graph[T, I, V], from, to T, weight I, isNonDir, isNonCyc bool) error {
	if g == nil {
		return fmt.Errorf("unable to read graph (nil): %w", errs.DoesNotExist)
	}
	m := g.Map()
	n := *m

	k := getKeysFromList(g)

	fromNode, ok := k[from]
	if !ok {
		return fmt.Errorf("from node: %w", errs.DoesNotExist)
	}
	toNode, ok := k[to]
	if !ok {
		return fmt.Errorf("to node: %w", errs.DoesNotExist)
	}

	if isNonCyc {
		ok, err := DepthFirstSearch(g, VerifyCycles(fromNode, toNode), toNode)
		if err != nil {
			return err
		}
		if !ok {
			return errs.CyclicalEdge
		}
	}

	if isNonDir {
		AddEdgeInListBi(n, fromNode, toNode, weight)
	} else {
		AddEdgeInListUni(n, fromNode, toNode, weight)
	}

	m = &n
	return nil
}

func AddEdgeInListUni[T model.ID, I model.Int, V any](m map[model.Node[T, I, V]][]model.Node[T, I, V], from, to model.Node[T, I, V], weight I) {
	m[from] = append(m[from], to)
}

func AddEdgeInListBi[T model.ID, I model.Int, V any](m map[model.Node[T, I, V]][]model.Node[T, I, V], from, to model.Node[T, I, V], weight I) {
	m[from] = append(m[from], to)
	m[to] = append(m[to], from)
}

func GetEdgesFromListNode[T model.ID, I model.Int, V any](g Graph[T, I, V], node T) ([]model.Node[T, I, V], error) {
	m := *g.Map()
	k := getKeysFromList(g)

	target, ok := k[node]
	if !ok {
		return nil, errs.DoesNotExist
	}

	return m[target], nil
}

func GetWeightFromEdgesInList[T model.ID, I model.Int, V any](g Graph[T, I, V], from, to T) (I, error) {
	fromNode, err := g.Get(from)
	if err != nil {
		return 0, err
	}

	toNode, err := g.Get(to)
	if err != nil {
		return 0, err
	}

	m := *g.Map()

	for _, v := range m[fromNode] {
		if v == toNode {
			return 1, nil
		}
	}

	return 0, nil
}

type output[T model.ID, I model.Int, V any] struct {
	ID    T         `json:"id"`
	Nodes map[T][]T `json:"nodes,omitempty"`
}

func (g *listGraph[T, I, V]) String() string {
	var out = output[T, I, V]{
		ID:    g.ID(),
		Nodes: map[T][]T{},
	}

	for ko, vo := range g.n {
		innerEdges := []T{}
		for _, ie := range vo {
			innerEdges = append(innerEdges, ie.ID())
		}
		out.Nodes[ko.ID()] = innerEdges
	}

	b, _ := json.Marshal(out)
	return string(b)
}
