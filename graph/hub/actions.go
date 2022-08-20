package hub

import (
	"encoding/json"
	"fmt"

	"github.com/zalgonoise/x/graph/errs"
	"github.com/zalgonoise/x/graph/model"
)

func getKeysFromMap[T model.ID, I model.Int](g model.Hub[T, I]) map[T]model.Hub[T, I] {
	m := *g.Map()
	keyMap := map[T]model.Hub[T, I]{}

	for k := range m {
		keyMap[k.ID()] = k
	}
	return keyMap
}

func AddNodesToMap[T model.ID, I model.Int](g model.Hub[T, I], nodes ...model.Hub[T, I]) error {
	m := g.Map()
	n := *m

	curKeys := getKeysFromMap(g)

	for _, node := range nodes {

		if _, ok := n[node]; ok {
			return errs.AlreadyExists
		}

		n[node] = map[model.Hub[T, I]]I{
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

func RemoveNodesFromMap[T model.ID, I model.Int](g model.Hub[T, I], ids ...T) error {
	m := g.Map()
	n := *m

	curKeys := getKeysFromMap(g)

	for _, id := range ids {
		node, err := g.Get(id)
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

func GetNodeFromMap[T model.ID, I model.Int](g model.Hub[T, I], node T) (model.Hub[T, I], error) {
	k := getKeysFromMap(g)

	n, ok := k[node]
	if !ok {
		return nil, errs.DoesNotExist
	}

	return n, nil
}

func GetKeysFromMap[T model.ID, I model.Int](g model.Hub[T, I]) ([]model.Hub[T, I], error) {
	m := *g.Map()

	out := []model.Hub[T, I]{}

	for k := range m {
		out = append(out, k)
	}

	return out, nil
}

func AddEdgeInMap[T model.ID, I model.Int](g model.Hub[T, I], from, to T, weight I, isNonDir, isNonCyc bool) error {
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

func AddEdgeInMapUni[T model.ID, I model.Int](m map[model.Hub[T, I]]map[model.Hub[T, I]]I, from, to model.Hub[T, I], weight I) {
	m[from][to] = weight
}

func AddEdgeInMapBi[T model.ID, I model.Int](m map[model.Hub[T, I]]map[model.Hub[T, I]]I, from, to model.Hub[T, I], weight I) {
	m[from][to] = weight
	m[to][from] = weight
}

func GetEdgesFromMapNode[T model.ID, I model.Int](g model.Hub[T, I], node T) ([]model.Hub[T, I], error) {
	var out []model.Hub[T, I]

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

func GetWeightFromEdgesInMap[T model.ID, I model.Int](g model.Hub[T, I], from, to T) (I, error) {
	fromNode, err := g.Get(from)
	if err != nil {
		return 0, err
	}

	toNode, err := g.Get(to)
	if err != nil {
		return 0, err
	}

	m := *g.Map()

	return m[fromNode][toNode], nil
}

func GetParentFromNode[T model.ID, I model.Int](g model.Hub[T, I], node T) (model.Hub[T, I], error) {
	n, err := g.Get(node)
	if err != nil {
		return nil, err
	}

	return n.Parent(), nil
}

type output[T model.ID, I model.Int] struct {
	ID    T             `json:"id"`
	Data  any           `json:"data,omitempty"`
	Nodes map[T]map[T]I `json:"nodes,omitempty"`
}

func (g *hubGraph[T, I]) String() string {
	var out = output[T, I]{
		ID:    g.ID(),
		Data:  g.Value(),
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
