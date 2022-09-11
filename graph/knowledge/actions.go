package knowledge

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

		n[node] = map[I][]model.Graph[T, I]{}

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
			for _, iEdges := range innerEdges {
				for _, e := range iEdges {
					if e.ID() == node.ID() {
						err := g.Disconnect(innerNode.ID(), node.ID())
						if err != nil {
							return err
						}
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

	if isNonCyc {
		ok, err := actions.DepthFirstSearch[T, I](g, actions.VerifyCycles(fromNode, toNode), toNode)
		if err != nil {
			return err
		}
		if !ok {
			return errs.CyclicalEdge
		}
	}

	if weight == 0 {
		if isNonDir {
			return RemoveEdgeFromList(fromNode, toNode)
		} else {
			err := RemoveEdgeFromList(fromNode, toNode)
			if err != nil {
				return fmt.Errorf("Error removing edge %v from node %v: %v", toNode.ID(), fromNode.ID(), err)
			}
			return RemoveEdgeFromList(toNode, fromNode)
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
	im := m[from]
	for prop, edges := range im {
		for idx, e := range edges {
			if e.ID() == to.ID() {
				edges[idx] = edges[len(edges)-1]
				im[prop] = edges[:len(edges)-1]
			}
		}
	}
	return nil
}

func AddEdgeInListUni[T model.ID, I model.Num](from, to model.Graph[T, I], weight I) error {
	g, ok := from.Parent().(Graph[T, I])
	if !ok {
		err := g.Connect(from.ID(), to.ID(), weight)
		if err != nil {
			return fmt.Errorf("node %v's parent graph %v does not support cross-graph connections: %w", to.ID(), g.ID(), err)
		}
		return nil
	}

	m := *g.adjacency()
	m[from][weight] = append(m[from][weight], to)
	return nil
}

func GetEdgesFromListNode[T model.ID, I model.Num](g Graph[T, I], node T) ([]model.Graph[T, I], error) {
	m := *g.adjacency()
	k := getKeysFromList(g)

	target, ok := k[node]
	if !ok {
		return nil, errs.DoesNotExist
	}

	var allEdges []model.Graph[T, I]

	for _, edges := range m[target] {
		for _, e := range edges {
			allEdges = append(allEdges, e)
		}
	}

	return allEdges, nil
}

func GetWeightFromEdgesInList[T model.ID, I model.Num](g Graph[T, I], from, to T) (I, error) {
	fromNode, err := g.Get(from)
	if err != nil {
		return 0, err
	}
	m := *g.adjacency()

	for prop, edges := range m[fromNode] {
		for _, e := range edges {
			if e.ID() == to {
				return prop, nil
			}
		}
	}

	return 0, nil
}

func GetEdgesWithProperty[T model.ID, I model.Num](g Graph[T, I], from T, weight I) ([]model.Graph[T, I], error) {
	fromNode, err := g.Get(from)
	if err != nil {
		return nil, err
	}
	m := *g.adjacency()

	var out []model.Graph[T, I]

	list, ok := m[fromNode][weight]
	if !ok {
		return out, nil
	}
	return list, nil
}

func GetNodeProperties[T model.ID, I model.Num](g Graph[T, I], from T) ([]I, error) {
	fromNode, err := g.Get(from)
	if err != nil {
		return nil, err
	}
	m := *g.adjacency()

	var out []I

	emap := m[fromNode]
	for k := range emap {
		out = append(out, k)
	}

	return out, nil
}

func (g *knowledgeGraph[T, I]) String() string {
	var dirSetting dot.Direction

	if g.conf.IsNonDirectional {
		dirSetting = dot.Undirected
	} else {
		dirSetting = dot.Directed
	}

	dotGraph := dot.New[T, I](dirSetting)

	for k, v := range g.n {
		for prop, ies := range v {
			for _, e := range ies {
				dotGraph.Add(k.ID(), e.ID(), prop)
			}
		}
	}
	return dotGraph.String()
}
