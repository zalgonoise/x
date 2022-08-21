package matrix

import (
	"encoding/json"
	"fmt"

	"github.com/zalgonoise/x/graph/actions"
	"github.com/zalgonoise/x/graph/errs"
	"github.com/zalgonoise/x/graph/model"
	"github.com/zalgonoise/x/graph/options"
)

func getKeysFromMap[T model.ID, I model.Num](g Graph[T, I]) map[T]model.Graph[T, I] {
	m := *g.adjancy()
	keyMap := map[T]model.Graph[T, I]{}

	for k := range m {
		keyMap[k.ID()] = k
	}
	return keyMap
}

func AddNodesToMap[T model.ID, I model.Num](g Graph[T, I], conf *options.GraphConfig, nodes ...model.Graph[T, I]) error {
	if conf.MaxDepth > 0 && actions.GraphDepth[T, I](g) >= conf.MaxDepth {
		return errs.MaxDepthReached
	}

	m := g.adjancy()
	n := *m

	count := len(n)

	for idx, node := range nodes {
		if conf.MaxNodes > 0 && count+idx >= conf.MaxNodes {
			return errs.MaxNodesReached
		}

		if _, ok := n[node]; ok {
			return errs.AlreadyExists
		}

		n[node] = map[model.Graph[T, I]]I{}

		for k := range n {
			// map this node to existing ones
			n[k][node] = 0

			// map other nodes to this node
			n[node][k] = 0
		}

		// link node to graph
		node.Link(g)
	}

	m = &n
	return nil
}

func RemoveNodesFromMap[T model.ID, I model.Num](g Graph[T, I], ids ...T) error {
	m := g.adjancy()
	n := *m

	for _, id := range ids {
		node, err := g.Get(id)
		if err != nil {
			return err
		}

		if _, ok := n[node]; !ok {
			return err
		}

		for k := range n {
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

func GetNodeFromMap[T model.ID, I model.Num](g Graph[T, I], node T) (model.Graph[T, I], error) {
	k := getKeysFromMap(g)

	n, ok := k[node]
	if !ok {
		return nil, errs.DoesNotExist
	}

	return n, nil
}

func ListNodesFromMap[T model.ID, I model.Num](g Graph[T, I]) ([]model.Graph[T, I], error) {
	m := *g.adjancy()

	out := []model.Graph[T, I]{}

	for k := range m {
		out = append(out, k)
	}

	return out, nil
}

func AddEdgeInMap[T model.ID, I model.Num](g Graph[T, I], from, to T, weight I, isNonDir, isNonCyc bool) error {
	m := g.adjancy()
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
		ok, err := actions.DepthFirstSearch[T, I](g, actions.VerifyCycles(fromNode, toNode), toNode)
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

func AddEdgeInMapUni[T model.ID, I model.Num](m map[model.Graph[T, I]]map[model.Graph[T, I]]I, from, to model.Graph[T, I], weight I) {
	m[from][to] = weight
}

func AddEdgeInMapBi[T model.ID, I model.Num](m map[model.Graph[T, I]]map[model.Graph[T, I]]I, from, to model.Graph[T, I], weight I) {
	m[from][to] = weight
	m[to][from] = weight
}

func GetEdgesFromMapNode[T model.ID, I model.Num](g Graph[T, I], node T) ([]model.Graph[T, I], error) {
	var out []model.Graph[T, I]

	m := *g.adjancy()
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

func GetWeightFromEdgesInMap[T model.ID, I model.Num](g Graph[T, I], from, to T) (I, error) {
	fromNode, err := g.Get(from)
	if err != nil {
		return 0, err
	}

	toNode, err := g.Get(to)
	if err != nil {
		return 0, err
	}

	m := *g.adjancy()

	return m[fromNode][toNode], nil
}

type output[T model.ID, I model.Num] struct {
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
