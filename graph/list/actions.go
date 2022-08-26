package list

import (
	"fmt"

	"github.com/zalgonoise/x/graph/actions"
	"github.com/zalgonoise/x/graph/dot"
	"github.com/zalgonoise/x/graph/errs"
	"github.com/zalgonoise/x/graph/model"
	"github.com/zalgonoise/x/graph/options"
)

func getKeysFromList[T model.ID, I model.Num](g Graph[T, I]) map[T]model.Graph[T, I] {
	m := *g.adjancy()
	keyMap := map[T]model.Graph[T, I]{}

	for k := range m {
		keyMap[k.ID()] = k
	}
	return keyMap
}

func AddNodesToList[T model.ID, I model.Num](g Graph[T, I], conf *options.GraphConfig, nodes ...model.Graph[T, I]) error {
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

		n[node] = []model.Graph[T, I]{}

		// link node to graph
		node.Link(g)
	}

	m = &n
	return nil
}

func RemoveNodesFromList[T model.ID, I model.Num](g Graph[T, I], ids ...T) error {
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

func GetNodeFromList[T model.ID, I model.Num](g Graph[T, I], node T) (model.Graph[T, I], error) {
	k := getKeysFromList(g)

	n, ok := k[node]
	if !ok {
		return nil, errs.DoesNotExist
	}

	return n, nil
}

func ListNodesFromList[T model.ID, I model.Num](g Graph[T, I]) ([]model.Graph[T, I], error) {
	m := *g.adjancy()

	out := []model.Graph[T, I]{}

	for k := range m {
		out = append(out, k)
	}

	return out, nil
}

func AddEdgeInList[T model.ID, I model.Num](g Graph[T, I], from, to T, weight I, isNonDir, isNonCyc bool) error {
	if g == nil {
		return fmt.Errorf("unable to read graph (nil): %w", errs.DoesNotExist)
	}
	m := g.adjancy()
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
		ok, err := actions.DepthFirstSearch[T, I](g, actions.VerifyCycles(fromNode, toNode), toNode)
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

func AddEdgeInListUni[T model.ID, I model.Num](m map[model.Graph[T, I]][]model.Graph[T, I], from, to model.Graph[T, I], weight I) {
	m[from] = append(m[from], &listEdge[T, I]{Graph: to, weight: weight})
}

func AddEdgeInListBi[T model.ID, I model.Num](m map[model.Graph[T, I]][]model.Graph[T, I], from, to model.Graph[T, I], weight I) {
	m[from] = append(m[from], &listEdge[T, I]{Graph: to, weight: weight})
	m[to] = append(m[to], &listEdge[T, I]{Graph: from, weight: weight})
}

func GetEdgesFromListNode[T model.ID, I model.Num](g Graph[T, I], node T) ([]model.Graph[T, I], error) {
	m := *g.adjancy()
	k := getKeysFromList(g)

	target, ok := k[node]
	if !ok {
		return nil, errs.DoesNotExist
	}

	return m[target], nil
}

func GetWeightFromEdgesInList[T model.ID, I model.Num](g Graph[T, I], from, to T) (I, error) {
	fromNode, err := g.Get(from)
	if err != nil {
		return 0, err
	}

	toNode, err := g.Get(to)
	if err != nil {
		return 0, err
	}

	m := *g.adjancy()

	for _, v := range m[fromNode] {
		if v == toNode {
			lnode, ok := v.(*listEdge[T, I])
			if !ok {
				return 1, nil
			}
			return lnode.weight, nil
		}
	}

	return 0, nil
}

type output[T model.ID, I model.Num] struct {
	ID    T         `json:"id"`
	Nodes map[T][]T `json:"nodes,omitempty"`
}

func (g *listGraph[T, I]) String() string {
	var dirSetting dot.Direction

	if g.conf.IsNonDirectional {
		dirSetting = dot.Undirected
	} else {
		dirSetting = dot.Directed
	}

	dotGraph := dot.New[T, I](dirSetting)

	for k, v := range g.n {
		for _, ie := range v {
			dotGraph.Add(k.ID(), ie.ID(), 1)
		}
	}
	return dotGraph.String()

	// var out = output[T, I]{
	// 	ID:    g.ID(),
	// 	Nodes: map[T][]T{},
	// }

	// for ko, vo := range g.n {
	// 	innerEdges := []T{}
	// 	for _, ie := range vo {
	// 		innerEdges = append(innerEdges, ie.ID())
	// 	}
	// 	out.Nodes[ko.ID()] = innerEdges
	// }

	// b, _ := json.Marshal(out)
	// return string(b)
}
