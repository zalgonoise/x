package list

import (
	"errors"
	"fmt"

	"github.com/zalgonoise/x/graph/actions"
	"github.com/zalgonoise/x/graph/dot"
	"github.com/zalgonoise/x/graph/errs"
	"github.com/zalgonoise/x/graph/model"
	"github.com/zalgonoise/x/graph/options"
)

func getKeysFromList[T model.ID, I model.Num](g Graph[T, I]) map[T]model.Graph[T, I] {
	m := *g.adjacency()
	keyMap := map[T]model.Graph[T, I]{}

	for k := range m {
		keyMap[k.ID()] = k
	}
	return keyMap
}

func AddNodesToList[T model.ID, I model.Num](g Graph[T, I], conf options.Setting, nodes ...model.Graph[T, I]) error {
	config := options.New(conf)

	if config.MaxDepth > 0 && actions.GraphDepth[T, I](g) >= config.MaxDepth {
		return errs.MaxDepthReached
	}

	m := g.adjacency()
	n := *m

	count := len(n)

	for idx, node := range nodes {
		if config.MaxNodes > 0 && count+idx >= config.MaxNodes {
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

		n[node] = []model.Graph[T, I]{}

	}

	*m = n
	return nil
}

func RemoveNodesFromList[T model.ID, I model.Num](g Graph[T, I], ids ...T) error {
	m := g.adjacency()
	n := *m

	for _, id := range ids {
		node, err := g.Get(id)
		if err != nil {
			return err
		}

		// disconnect any edges
		for innerNode, innerEdges := range n {
			if innerNode == node {
				continue
			}
			for _, e := range innerEdges {
				if e.ID() == node.ID() {
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
		err = node.Link(nil)
		if err != nil {
			return err
		}
	}

	*m = n
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
	m := *g.adjacency()

	out := []model.Graph[T, I]{}

	for k := range m {
		out = append(out, k)
	}

	return out, nil
}

func AddEdgeInList[T model.ID, I model.Num](g Graph[T, I], from, to T, weight I, isNonDir, isNonCyc bool) error {
	var (
		graph  model.Graph[T, I] = g
		toNode model.Graph[T, I]
	)

	if g == nil {
		return fmt.Errorf("unable to read graph (nil): %w", errs.DoesNotExist)
	}
	k := getKeysFromList(g)

	fromNode, ok := k[from]
	if !ok {
		return fmt.Errorf("from node: %w", errs.DoesNotExist)
	}

	toNode, ok = k[to]
	if !ok {
		parent, err := actions.LeafLookup(graph, to)
		if err != nil && errors.Is(err, errs.DoesNotExist) {
			return fmt.Errorf("to node: %w", err)
		}
		toNode, err = parent.Get(to)
		if err != nil {
			return err
		}
	}

	if weight == 0 {
		if isNonDir {
			err := RemoveEdgeFromList(fromNode, toNode)
			if err != nil {
				return fmt.Errorf("Error removing edge %v from node %v: %v", toNode.ID(), fromNode.ID(), err)
			}
			return RemoveEdgeFromList(toNode, fromNode)
		}
		return RemoveEdgeFromList(fromNode, toNode)
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
		err := AddEdgeInListUni(fromNode, toNode, weight)
		if err != nil {
			return fmt.Errorf("Error adding edge %v from node %v with weight %v: %v", toNode.ID(), fromNode.ID(), weight, err)
		}
		return AddEdgeInListUni(toNode, fromNode, weight)
	}
	return AddEdgeInListUni(fromNode, toNode, weight)
}

func RemoveEdgeFromList[T model.ID, I model.Num](from, to model.Graph[T, I]) error {
	g, ok := from.Parent().(Graph[T, I])
	if !ok {
		err := g.Disconnect(from.ID(), to.ID())
		if err != nil {
			return fmt.Errorf("node %v's parent graph %v does not support cross-graph connections: %w", to.ID(), g.ID(), err)
		}
		return nil
	}

	m := *g.adjacency()
	list := m[from]
	for idx, e := range list {
		if e.ID() == to.ID() {
			list[idx] = list[len(list)-1]
			m[from] = list[:len(list)-1]
		}
	}
	return nil
}

func AddEdgeInListUni[T model.ID, I model.Num](from, to model.Graph[T, I], weight I) error {
	g, ok := from.Parent().(Graph[T, I])
	if !ok {
		err := from.Parent().Connect(from.ID(), to.ID(), weight)
		if err != nil {
			return fmt.Errorf("node %v's parent graph %v does not support cross-graph connections: %w", to.ID(), from.Parent().ID(), err)
		}
		return nil
	}

	m := *g.adjacency()
	if weight == 1 {
		m[from] = append(m[from], to)
		return nil
	}

	m[from] = append(m[from], &listEdge[T, I]{Graph: to, weight: weight})
	return nil
}

func GetEdgesFromListNode[T model.ID, I model.Num](g Graph[T, I], node T) ([]model.Graph[T, I], error) {
	m := *g.adjacency()
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
	m := *g.adjacency()

	for _, v := range m[fromNode] {
		if v.ID() == to {
			lnode, ok := v.(*listEdge[T, I])
			if !ok {
				return 1, nil
			}
			return lnode.weight, nil
		}
	}

	return 0, nil
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
}
