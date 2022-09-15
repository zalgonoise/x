package matrix

import (
	"fmt"

	"github.com/zalgonoise/x/graph/actions"
	"github.com/zalgonoise/x/graph/dot"
	"github.com/zalgonoise/x/graph/errs"
	"github.com/zalgonoise/x/graph/model"
	"github.com/zalgonoise/x/graph/options"
)

func getKeysFromMap[T model.ID, I model.Num](g Graph[T, I]) map[T]model.Graph[T, I] {
	m := *g.adjacency()
	keyMap := map[T]model.Graph[T, I]{}

	for k := range m {
		keyMap[k.ID()] = k
	}
	return keyMap
}

func AddNodesToMap[T model.ID, I model.Num](g Graph[T, I], config options.Setting, nodes ...model.Graph[T, I]) error {
	conf := options.New(config)
	if conf.MaxDepth > 0 && actions.GraphDepth[T, I](g) >= conf.MaxDepth {
		return errs.MaxDepthReached
	}

	m := g.adjacency()
	n := *m

	count := len(n)

	for idx, node := range nodes {
		if conf.MaxNodes > 0 && count+idx >= conf.MaxNodes {
			return errs.MaxNodesReached
		}

		if _, ok := n[node]; ok {
			return errs.AlreadyExists
		}

		// link node to parent before adding it to graph
		err := node.Link(g, conf)
		if err != nil {
			return err
		}

		n[node] = map[model.Graph[T, I]]I{}

		for k := range n {
			// map this node to existing ones
			n[k][node] = 0

			// map other nodes to this node
			n[node][k] = 0
		}

	}

	*m = n
	return nil
}

func RemoveNodesFromMap[T model.ID, I model.Num](g Graph[T, I], ids ...T) error {
	m := g.adjacency()
	n := *m

	for _, id := range ids {
		node, err := g.Get(id)
		if err != nil {
			return err
		}

		if _, ok := n[node]; !ok {
			return errs.DoesNotExist
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
		err = node.Link(nil)
		if err != nil {
			return err
		}
	}

	*m = n
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
	m := *g.adjacency()

	out := []model.Graph[T, I]{}

	for k := range m {
		out = append(out, k)
	}

	return out, nil
}

func AddEdgeInMap[T model.ID, I model.Num](g Graph[T, I], from, to T, weight I, isNonDir, isNonCyc bool) error {
	m := g.adjacency()
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
		setEdgeInMap(n, fromNode, toNode, weight)
		setEdgeInMap(n, toNode, fromNode, weight)
	} else {
		setEdgeInMap(n, fromNode, toNode, weight)
	}

	*m = n
	return nil
}

func setEdgeInMap[T model.ID, I model.Num](m map[model.Graph[T, I]]map[model.Graph[T, I]]I, from, to model.Graph[T, I], weight I) {
	m[from][to] = weight
}

func GetEdgesFromMapNode[T model.ID, I model.Num](g Graph[T, I], node T) ([]model.Graph[T, I], error) {
	var out []model.Graph[T, I]

	m := *g.adjacency()
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

	m := *g.adjacency()

	return m[fromNode][toNode], nil
}

func (g *mapGraph[T, I]) String() string {
	var dirSetting dot.Direction

	if g.conf.IsNonDirectional {
		dirSetting = dot.Undirected
	} else {
		dirSetting = dot.Directed
	}

	dotGraph := dot.New[T, I](dirSetting)

	for k, v := range g.n {
		for ki, vi := range v {
			dotGraph.Add(k.ID(), ki.ID(), vi)
		}
	}
	return dotGraph.String()
}
