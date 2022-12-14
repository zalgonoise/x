package linkedlist

import (
	"fmt"

	"github.com/zalgonoise/x/graph/dot"
	"github.com/zalgonoise/x/graph/errs"
	"github.com/zalgonoise/x/graph/model"
	"github.com/zalgonoise/x/graph/options"
)

func GetGraphMap[T model.ID, I model.Num](g Graph[T, I]) ([]model.Graph[T, I], error) {
	node, ok := GetFirstNode(g).(Graph[T, I])
	if !ok {
		return nil, errs.InvalidType
	}
	nodes := []model.Graph[T, I]{0: node}
	idx := 1

	for node.nextGraph(nil) != nil {
		next, ok := node.nextGraph(nil).(Graph[T, I])
		if !ok {
			return nil, errs.InvalidType
		}
		nodes[idx] = node
		idx++
		node = next
	}

	return nodes, nil
}

func GetLastNode[T model.ID, I model.Num](g Graph[T, I]) model.Graph[T, I] {
	graph := g.(*linkedList[T, I])

	for graph.nextGraph(nil) != nil {
		graph = graph.nextGraph(nil).(*linkedList[T, I])
	}
	return graph
}
func GetFirstNode[T model.ID, I model.Num](g Graph[T, I]) model.Graph[T, I] {
	graph := g.(*linkedList[T, I])

	for graph.parentGraph(nil) != nil {
		graph = graph.parentGraph(nil).(*linkedList[T, I])
	}
	return graph
}

func AddNodesToList[T model.ID, I model.Num](g Graph[T, I], conf options.Setting, nodes ...model.Graph[T, I]) error {
	last, ok := GetLastNode(g).(Graph[T, I])
	if !ok {
		return errs.InvalidType
	}

	for _, node := range nodes {
		n, ok := node.(Graph[T, I])
		if !ok {
			return errs.InvalidType
		}

		last.nextGraph(node)
		last = n
	}

	return nil
}

func RemoveNodesFromList[T model.ID, I model.Num](g Graph[T, I], ids ...T) error {
	all, err := GetGraphMap(g)
	if err != nil {
		return err
	}

	lastIdx := len(all) - 1

	for _, input := range ids {
		for k, v := range all {
			if input == v.ID() {
				if k == 0 {
					modifier, ok := all[1].(*linkedList[T, I])
					if !ok {
						return errs.InvalidType
					}
					modifier.parent = nil
					continue
				}
				if k == lastIdx {
					modifier, ok := all[k-1].(*linkedList[T, I])
					if !ok {
						return errs.InvalidType
					}
					modifier.parent = nil
					continue
				}

				parent, ok := all[k-1].(*linkedList[T, I])
				if !ok {
					return fmt.Errorf("failed to get parent: %w", errs.InvalidType)
				}
				next, ok := all[k+1].(*linkedList[T, I])
				if !ok {
					return fmt.Errorf("failed to get next: %w", errs.InvalidType)
				}
				parent.nextGraph(next)
				next.parentGraph(parent)
			}
		}
	}

	return nil
}

func GetNodeFromList[T model.ID, I model.Num](g Graph[T, I], node T) (model.Graph[T, I], error) {
	all, err := GetGraphMap(g)
	if err != nil {
		return nil, err
	}

	for _, v := range all {
		if v.ID() == node {
			return v, nil
		}
	}

	return nil, errs.DoesNotExist
}

func ListNodesFromList[T model.ID, I model.Num](g Graph[T, I]) ([]model.Graph[T, I], error) {

	all, err := GetGraphMap(g)
	if err != nil {
		return nil, err
	}
	return all, nil
}

type output[T model.ID, I model.Num] struct {
	ID    T         `json:"id"`
	Nodes map[int]T `json:"nodes,omitempty"`
}

func (g *linkedList[T, I]) String() string {
	var dirSetting dot.Direction

	if g.conf.IsNonDirectional {
		dirSetting = dot.Undirected
	} else {
		dirSetting = dot.Directed
	}

	dotGraph := dot.New[T, I](dirSetting)

	all, _ := GetGraphMap[T, I](g)
	for i := 1; i < len(all); i++ {
		dotGraph.Add(all[i-1].ID(), all[i].ID(), 1)
	}
	return dotGraph.String()

	// var out = output[T, I]{
	// 	ID:    g.ID(),
	// 	Nodes: map[int]T{},
	// }

	// all, _ := GetGraphMap[T, I](g)
	// for k, v := range all {
	// 	out.Nodes[k] = v.ID()
	// }

	// b, _ := json.Marshal(out)
	// return string(b)
}
