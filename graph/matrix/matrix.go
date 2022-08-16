package matrix

import (
	"encoding/json"
	"fmt"

	"github.com/zalgonoise/x/graph"
)

type Graph[T graph.ID, I graph.Int] interface {
	graph.Graph[T, I]
	Map() *map[graph.Node[T]]map[graph.Node[T]]I
	Keys() *map[T]graph.Node[T]
}

type mapNode[T graph.ID] struct {
	id T
}

func (n *mapNode[T]) ID() T {
	return n.id
}

func NewNode[T graph.ID](id T) graph.Node[T] {
	return &mapNode[T]{id: id}
}

type mapGraph[T graph.ID, I graph.Int] struct {
	id   T
	n    map[graph.Node[T]]map[graph.Node[T]]I
	keys map[T]graph.Node[T]
}

func NewGraph[T graph.ID, I graph.Int](id T) graph.Graph[T, I] {
	return &mapGraph[T, I]{
		id:   id,
		n:    map[graph.Node[T]]map[graph.Node[T]]I{},
		keys: map[T]graph.Node[T]{},
	}
}

func (g *mapGraph[T, I]) Map() *map[graph.Node[T]]map[graph.Node[T]]I {
	return &g.n
}
func (g *mapGraph[T, I]) Keys() *map[T]graph.Node[T] {
	return &g.keys
}
func (g *mapGraph[T, I]) ID() T {
	return g.id
}
func (g *mapGraph[T, I]) AddNode(nodes ...graph.Node[T]) error {
	return AddNodesToMap[T, I](g, nodes...)
}
func (g *mapGraph[T, I]) RemoveNode(nodes ...T) error {
	return RemoveNodesFromMap[T, I](g, nodes...)
}
func (g *mapGraph[T, I]) GetNode(node T) (graph.Node[T], error) {
	return GetNodeFromMap[T, I](g, node)
}
func (g *mapGraph[T, I]) Get() ([]graph.Node[T], error) {
	return GetKeysFromMap[T, I](g)
}
func (g *mapGraph[T, I]) AddEdge(from, to T, weight I) error {
	return AddEdgeInMapUni[T, I](g, from, to, weight)
}
func (g *mapGraph[T, I]) RemoveEdge(target, edge T) error {
	return AddEdgeInMapUni[T, I](g, target, edge, 0)
}
func (g *mapGraph[T, I]) GetEdges(node T) ([]graph.Node[T], error) {
	return GetEdgesFromMapNode[T, I](g, node)
}

func AddNodesToMap[T graph.ID, I graph.Int](g Graph[T, I], nodes ...graph.Node[T]) error {
	m := g.Map()
	n := *m

	c := g.Keys()
	curKeys := *c

	for _, node := range nodes {

		if _, ok := n[node]; ok {
			return graph.ErrAlreadyExists
		}

		n[node] = map[graph.Node[T]]I{
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

func RemoveNodesFromMap[T graph.ID, I graph.Int](g Graph[T, I], ids ...T) error {
	m := g.Map()
	n := *m

	curKeys := make([]graph.Node[T], len(n))
	for k := range n {
		curKeys = append(curKeys, k)
	}

	for _, id := range ids {
		node, err := g.GetNode(id)
		if err != nil {
			return err
		}

		if _, ok := n[node]; !ok {
			return graph.ErrDoesNotExist
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

func GetNodeFromMap[T graph.ID, I graph.Int](g Graph[T, I], node T) (graph.Node[T], error) {
	k := *g.Keys()

	n, ok := k[node]
	if !ok {
		return nil, graph.ErrDoesNotExist
	}

	return n, nil
}

func GetKeysFromMap[T graph.ID, I graph.Int](g Graph[T, I]) ([]graph.Node[T], error) {
	m := *g.Map()

	out := make([]graph.Node[T], len(m))

	for k := range m {
		out = append(out, k)
	}

	return out, nil
}

func AddEdgeInMapUni[T graph.ID, I graph.Int](g Graph[T, I], from, to T, weight I) error {
	m := g.Map()
	n := *m

	k := *g.Keys()

	fromNode, ok := k[from]
	if !ok {
		return fmt.Errorf("from node: %w", graph.ErrDoesNotExist)
	}
	toNode, ok := k[to]
	if !ok {
		return fmt.Errorf("to node: %w", graph.ErrDoesNotExist)
	}

	n[fromNode][toNode] = weight

	m = &n
	return nil
}

func AddEdgeInMapBi[T graph.ID, I graph.Int](g Graph[T, I], from, to T, weight I) error {
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

func GetEdgesFromMapNode[T graph.ID, I graph.Int](g Graph[T, I], node T) ([]graph.Node[T], error) {
	var out []graph.Node[T]

	m := *g.Map()
	k := *g.Keys()

	target, ok := k[node]
	if !ok {
		return nil, graph.ErrDoesNotExist
	}

	conn, ok := m[target]
	if !ok {
		return nil, graph.ErrDoesNotExist
	}

	for k, v := range conn {
		if v == 0 || k.ID() == node {
			continue
		}
		out = append(out, k)
	}

	return out, nil

}

type output[T graph.ID, I graph.Int] struct {
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
