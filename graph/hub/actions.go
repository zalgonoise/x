package hub

import (
	"fmt"

	"github.com/zalgonoise/x/graph/errs"
	"github.com/zalgonoise/x/graph/model"
)

func getKeysFromMap[T model.ID, I model.Int, V any](g model.Hub[T, I, V]) map[T]model.Hub[T, I, V] {
	m := *g.Map()
	keyMap := map[T]model.Hub[T, I, V]{}

	for k := range m {
		keyMap[k.ID()] = k
	}
	return keyMap
}

func AddNodesToMap[T model.ID, I model.Int, V any](g model.Hub[T, I, V], nodes ...model.Hub[T, I, V]) error {
	m := g.Map()
	n := *m

	curKeys := getKeysFromMap(g)

	for _, node := range nodes {

		if _, ok := n[node]; ok {
			return errs.AlreadyExists
		}

		n[node] = map[model.Hub[T, I, V]]I{
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

func RemoveNodesFromMap[T model.ID, I model.Int, V any](g model.Hub[T, I, V], ids ...T) error {
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

func GetNodeFromMap[T model.ID, I model.Int, V any](g model.Hub[T, I, V], node T) (model.Hub[T, I, V], error) {
	k := getKeysFromMap(g)

	n, ok := k[node]
	if !ok {
		return nil, errs.DoesNotExist
	}

	return n, nil
}

func GetKeysFromMap[T model.ID, I model.Int, V any](g model.Hub[T, I, V]) ([]model.Hub[T, I, V], error) {
	m := *g.Map()

	out := []model.Hub[T, I, V]{}

	for k := range m {
		out = append(out, k)
	}

	return out, nil
}

func AddEdgeInMap[T model.ID, I model.Int, V any](g model.Hub[T, I, V], from, to T, weight I, isNonDir, isNonCyc bool) error {
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

func AddEdgeInMapUni[T model.ID, I model.Int, V any](m map[model.Hub[T, I, V]]map[model.Hub[T, I, V]]I, from, to model.Hub[T, I, V], weight I) {
	m[from][to] = weight
}

func AddEdgeInMapBi[T model.ID, I model.Int, V any](m map[model.Hub[T, I, V]]map[model.Hub[T, I, V]]I, from, to model.Hub[T, I, V], weight I) {
	m[from][to] = weight
	m[to][from] = weight
}

func GetEdgesFromMapNode[T model.ID, I model.Int, V any](g model.Hub[T, I, V], node T) ([]model.Hub[T, I, V], error) {
	var out []model.Hub[T, I, V]

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

func GetWeightFromEdgesInMap[T model.ID, I model.Int, V any](g model.Hub[T, I, V], from, to T) (I, error) {
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

func GetParentFromNode[T model.ID, I model.Int, V any](g model.Hub[T, I, V], node T) (model.Hub[T, I, V], error) {
	n, err := g.GetNode(node)
	if err != nil {
		return nil, err
	}

	return n.Parent(), nil
}
