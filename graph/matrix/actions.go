package matrix

import (
	"encoding/json"
	"fmt"

	"github.com/zalgonoise/x/graph/errs"
	"github.com/zalgonoise/x/graph/model"
)

func AddNodesToMap[T model.ID, I model.Int](g Graph[T, I], nodes ...model.Node[T]) error {
	m := g.Map()
	n := *m

	c := g.Keys()
	curKeys := *c

	for _, node := range nodes {

		if _, ok := n[node]; ok {
			return errs.AlreadyExists
		}

		n[node] = map[model.Node[T]]I{
			node: 0,
		}

		for _, k := range curKeys {
			// map this node to existing ones
			n[k][node] = 0

			// map other nodes to this node
			n[node][k] = 0
		}

		curKeys[node.ID()] = node
	}

	m = &n
	c = &curKeys
	return nil
}

func RemoveNodesFromMap[T model.ID, I model.Int](g Graph[T, I], ids ...T) error {
	m := g.Map()
	n := *m

	curKeys := make([]model.Node[T], len(n))
	for k := range n {
		curKeys = append(curKeys, k)
	}

	for _, id := range ids {
		node, err := g.GetNode(id)
		if err != nil {
			return err
		}

		if _, ok := n[node]; !ok {
			return err
		}

		for _, k := range curKeys {
			if k != node {
				delete(n[k], node)
				delete(n[node], k)
			}
		}

		delete(n[node], node)
		delete(n, node)
	}

	m = &n
	return nil
}

func GetNodeFromMap[T model.ID, I model.Int](g Graph[T, I], node T) (model.Node[T], error) {
	k := *g.Keys()

	n, ok := k[node]
	if !ok {
		return nil, errs.DoesNotExist
	}

	return n, nil
}

func GetKeysFromMap[T model.ID, I model.Int](g Graph[T, I]) ([]model.Node[T], error) {
	m := *g.Map()

	out := make([]model.Node[T], len(m))

	for k := range m {
		out = append(out, k)
	}

	return out, nil
}

func AddEdgeInMapUni[T model.ID, I model.Int](g Graph[T, I], from, to T, weight I) error {
	m := g.Map()
	n := *m

	k := *g.Keys()

	fromNode, ok := k[from]
	if !ok {
		return fmt.Errorf("from node: %w", errs.DoesNotExist)
	}
	toNode, ok := k[to]
	if !ok {
		return fmt.Errorf("to node: %w", errs.DoesNotExist)
	}

	n[fromNode][toNode] = weight

	m = &n
	return nil
}

func AddEdgeInMapBi[T model.ID, I model.Int](g Graph[T, I], from, to T, weight I) error {
	err := AddEdgeInMapUni(g, from, to, weight)
	if err != nil {
		return fmt.Errorf("creating from-to edge: %w", err)
	}

	err = AddEdgeInMapUni(g, to, from, weight)
	if err != nil {
		return fmt.Errorf("creating to-from edge: %w", err)
	}

	return nil
}

func GetEdgesFromMapNode[T model.ID, I model.Int](g Graph[T, I], node T) ([]model.Node[T], error) {
	var out []model.Node[T]

	m := *g.Map()
	k := *g.Keys()

	target, ok := k[node]
	if !ok {
		return nil, errs.DoesNotExist
	}

	conn, ok := m[target]
	if !ok {
		return nil, errs.DoesNotExist
	}

	for k, v := range conn {
		if v == 0 || k.ID() == node {
			continue
		}
		out = append(out, k)
	}

	return out, nil

}

type output[T model.ID, I model.Int] struct {
	ID    T             `json:"id"`
	Nodes map[T]map[T]I `json:"nodes,omitempty"`
}

func (g *mapGraph[T, I]) String() string {
	var out = output[T, I]{
		ID:    g.ID(),
		Nodes: map[T]map[T]I{},
	}

	for ko, vo := range g.n {
		innerMap := map[T]I{}
		for ki, vi := range vo {
			innerMap[ki.ID()] = vi
		}
		out.Nodes[ko.ID()] = innerMap
	}

	b, _ := json.Marshal(out)
	return string(b)
}
