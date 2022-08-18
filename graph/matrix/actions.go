package matrix

import (
	"encoding/json"
	"fmt"

	"github.com/zalgonoise/x/graph/errs"
	"github.com/zalgonoise/x/graph/model"
)

func getKeysFromMap[T model.ID, I model.Int, V any](g Graph[T, I, V]) map[T]model.Node[T, I, V] {
	m := *g.Map()
	keyMap := map[T]model.Node[T, I, V]{}

	for k := range m {
		keyMap[k.ID()] = k
	}
	return keyMap
}

func AddNodesToMap[T model.ID, I model.Int, V any](g Graph[T, I, V], nodes ...model.Node[T, I, V]) error {
	m := g.Map()
	n := *m

	curKeys := getKeysFromMap(g)

	for _, node := range nodes {

		if _, ok := n[node]; ok {
			return errs.AlreadyExists
		}

		n[node] = map[model.Node[T, I, V]]I{
			node: 0,
		}

		for _, k := range curKeys {
			// map this node to existing ones
			n[k][node] = 0

			// map other nodes to this node
			n[node][k] = 0
		}

		// link node to graph
		node.Link(g)

		// node appended to added keys
		curKeys[node.ID()] = node
	}

	m = &n
	return nil
}

func RemoveNodesFromMap[T model.ID, I model.Int, V any](g Graph[T, I, V], ids ...T) error {
	m := g.Map()
	n := *m

	curKeys := getKeysFromMap(g)

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

		// unlink node from graph
		node.Link(nil)
	}

	m = &n
	return nil
}

func GetNodeFromMap[T model.ID, I model.Int, V any](g Graph[T, I, V], node T) (model.Node[T, I, V], error) {
	k := getKeysFromMap(g)

	n, ok := k[node]
	if !ok {
		return nil, errs.DoesNotExist
	}

	return n, nil
}

func GetKeysFromMap[T model.ID, I model.Int, V any](g Graph[T, I, V]) ([]model.Node[T, I, V], error) {
	m := *g.Map()

	out := []model.Node[T, I, V]{}

	for k := range m {
		out = append(out, k)
	}

	return out, nil
}

func AddEdgeInMap[T model.ID, I model.Int, V any](g Graph[T, I, V], from, to T, weight I, isNonDir, isNonCyc bool) error {
	if g == nil {
		return fmt.Errorf("unable to read graph (nil): %w", errs.DoesNotExist)
	}
	m := g.Map()
	n := *m

	k := getKeysFromMap(g)

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
		AddEdgeInMapBi(n, fromNode, toNode, weight)
	} else {
		AddEdgeInMapUni(n, fromNode, toNode, weight)
	}

	m = &n
	return nil
}

func AddEdgeInMapUni[T model.ID, I model.Int, V any](m map[model.Node[T, I, V]]map[model.Node[T, I, V]]I, from, to model.Node[T, I, V], weight I) {
	m[from][to] = weight
}

func AddEdgeInMapBi[T model.ID, I model.Int, V any](m map[model.Node[T, I, V]]map[model.Node[T, I, V]]I, from, to model.Node[T, I, V], weight I) {
	m[from][to] = weight
	m[to][from] = weight
}

func GetEdgesFromMapNode[T model.ID, I model.Int, V any](g Graph[T, I, V], node T) ([]model.Node[T, I, V], error) {
	var out []model.Node[T, I, V]

	m := *g.Map()
	k := getKeysFromMap(g)

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

func GetWeightFromEdgesInMap[T model.ID, I model.Int, V any](g Graph[T, I, V], from, to T) (I, error) {
	fromNode, err := g.GetNode(from)
	if err != nil {
		return 0, err
	}

	toNode, err := g.GetNode(to)
	if err != nil {
		return 0, err
	}

	m := *g.Map()

	return m[fromNode][toNode], nil
}

func GetParentFromNode[T model.ID, I model.Int, V any](g Graph[T, I, V], node T) (model.Graph[T, I, V], error) {
	n, err := g.GetNode(node)
	if err != nil {
		return nil, err
	}

	return n.Parent(), nil
}

type output[T model.ID, I model.Int, V any] struct {
	ID    T             `json:"id"`
	Nodes map[T]map[T]I `json:"nodes,omitempty"`
}

func (g *mapGraph[T, I, V]) String() string {
	var out = output[T, I, V]{
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
