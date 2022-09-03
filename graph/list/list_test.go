package list

import (
	"errors"
	"reflect"
	"testing"

	"github.com/zalgonoise/x/graph/errs"
	"github.com/zalgonoise/x/graph/model"
	"github.com/zalgonoise/x/graph/options"
)

const (
	testIDString string  = "test-id"
	testIDInt    int     = 0
	testIDFloat  float32 = 0
)

type testObject struct {
	id       int
	name     string
	isActive bool
}
type ider interface {
	ID() int
}

func (o *testObject) ID() int {
	return o.id
}

func TestNew(t *testing.T) {
	t.Run("StringID", func(t *testing.T) {
		t.Run("IntWeight", func(t *testing.T) {
			t.Run("Success", func(t *testing.T) {
				g := New[string, int](testIDString, testIDString, options.GraphList)

				_, ok := g.(*listGraph[string, int])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *listGraph[string, int], got %T", g)
				}
			})
			t.Run("SuccessWithNilValue", func(t *testing.T) {
				g := New[string, int](testIDString, nil, options.GraphList)

				_, ok := g.(*listGraph[string, int])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *listGraph[string, int], got %T", g)
				}
			})
			t.Run("SuccessWithEmptyValue", func(t *testing.T) {
				g := New[string, int](testIDString, options.NoType, options.GraphList)

				_, ok := g.(*listGraph[string, int])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *listGraph[string, int], got %T", g)
				}
			})
			t.Run("SuccessWithInvalidOption", func(t *testing.T) {
				g := New[string, int](testIDString, options.NoType, options.GraphMatrix)

				_, ok := g.(*listGraph[string, int])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *listGraph[string, int], got %T", g)
				}
			})
			t.Run("SuccessWithNoConfig", func(t *testing.T) {
				g := New[string, int](testIDString, options.NoType, nil)

				_, ok := g.(*listGraph[string, int])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *listGraph[string, int], got %T", g)
				}
			})
		})
		t.Run("Float64Weight", func(t *testing.T) {
			t.Run("Success", func(t *testing.T) {
				g := New[string, float64](testIDString, testIDString, options.GraphList)

				_, ok := g.(*listGraph[string, float64])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *listGraph[string, int], got %T", g)
				}
			})
			t.Run("SuccessWithNilValue", func(t *testing.T) {
				g := New[string, float64](testIDString, nil, options.GraphList)

				_, ok := g.(*listGraph[string, float64])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *listGraph[string, int], got %T", g)
				}
			})
			t.Run("SuccessWithEmptyValue", func(t *testing.T) {
				g := New[string, float64](testIDString, options.NoType, options.GraphList)

				_, ok := g.(*listGraph[string, float64])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *listGraph[string, int], got %T", g)
				}
			})
		})

	})
	t.Run("IntID", func(t *testing.T) {
		t.Run("IntWeight", func(t *testing.T) {
			t.Run("Success", func(t *testing.T) {
				g := New[int, int](testIDInt, testIDInt, options.GraphList)

				_, ok := g.(*listGraph[int, int])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *listGraph[int, int], got %T", g)
				}
			})
			t.Run("SuccessWithNilValue", func(t *testing.T) {
				g := New[int, int](testIDInt, nil, options.GraphList)

				_, ok := g.(*listGraph[int, int])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *listGraph[int, int], got %T", g)
				}
			})
			t.Run("SuccessWithEmptyValue", func(t *testing.T) {
				g := New[int, int](testIDInt, options.NoType, options.GraphList)

				_, ok := g.(*listGraph[int, int])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *listGraph[int, int], got %T", g)
				}
			})
		})
		t.Run("Float64Weight", func(t *testing.T) {
			t.Run("Success", func(t *testing.T) {
				g := New[int, float64](testIDInt, testIDInt, options.GraphList)

				_, ok := g.(*listGraph[int, float64])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *listGraph[int, int], got %T", g)
				}
			})
			t.Run("SuccessWithNilValue", func(t *testing.T) {
				g := New[int, float64](testIDInt, nil, options.GraphList)

				_, ok := g.(*listGraph[int, float64])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *listGraph[int, int], got %T", g)
				}
			})
			t.Run("SuccessWithEmptyValue", func(t *testing.T) {
				g := New[int, float64](testIDInt, options.NoType, options.GraphList)

				_, ok := g.(*listGraph[int, float64])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *listGraph[int, int], got %T", g)
				}
			})
		})
	})
	t.Run("FloatID", func(t *testing.T) {
		t.Run("IntWeight", func(t *testing.T) {
			t.Run("Success", func(t *testing.T) {
				g := New[float32, int](testIDFloat, testIDFloat, options.GraphList)

				_, ok := g.(*listGraph[float32, int])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *listGraph[float32, int], got %T", g)
				}
			})
			t.Run("SuccessWithNilValue", func(t *testing.T) {
				g := New[float32, int](testIDFloat, nil, options.GraphList)

				_, ok := g.(*listGraph[float32, int])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *listGraph[float32, int], got %T", g)
				}
			})
			t.Run("SuccessWithEmptyValue", func(t *testing.T) {
				g := New[float32, int](testIDFloat, options.NoType, options.GraphList)

				_, ok := g.(*listGraph[float32, int])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *listGraph[float32, int], got %T", g)
				}
			})
		})
		t.Run("Float64Weight", func(t *testing.T) {
			t.Run("Success", func(t *testing.T) {
				g := New[float32, float64](testIDFloat, testIDFloat, options.GraphList)

				_, ok := g.(*listGraph[float32, float64])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *listGraph[float32, int], got %T", g)
				}
			})
			t.Run("SuccessWithNilValue", func(t *testing.T) {
				g := New[float32, float64](testIDFloat, nil, options.GraphList)

				_, ok := g.(*listGraph[float32, float64])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *listGraph[float32, int], got %T", g)
				}
			})
			t.Run("SuccessWithEmptyValue", func(t *testing.T) {
				g := New[float32, float64](testIDFloat, options.NoType, options.GraphList)

				_, ok := g.(*listGraph[float32, float64])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *listGraph[float32, int], got %T", g)
				}
			})
		})
	})
}

func TestAdjacency(t *testing.T) {
	root := New[string, int](testIDString, options.NoType, nil).(Graph[string, int])

	t.Run("EmptyAdjacencyList", func(t *testing.T) {
		m := root.adjacency()

		if len(*m) != 0 {
			t.Errorf("unexpected graph length -- wanted %v ; got %v", 0, len(*m))
		}
	})
	t.Run("OneNodeNoEdges", func(t *testing.T) {
		nodes := []model.Graph[string, int]{
			New[string, int]("alpha", options.NoType, nil),
		}
		var nodeIDs []string
		for _, n := range nodes {
			nodeIDs = append(nodeIDs, n.ID())
		}

		root.Add(nodes...)
		defer root.Remove(nodeIDs...)
		m := root.adjacency()

		if len(*m) != len(nodes) {
			t.Errorf("unexpected graph length -- wanted %v ; got %v", len(nodes), len(*m))
		}
		for k := range *m {
			var ok bool
			for _, id := range nodeIDs {
				if id == k.ID() {
					ok = true
					break
				}
				if !ok {
					t.Errorf("unable to find node %s in the adjacency list %v", k.ID(), nodeIDs)
				}
			}
		}
	})
	t.Run("ThreeNodesNoEdges", func(t *testing.T) {
		nodes := []model.Graph[string, int]{
			New[string, int]("alpha", options.NoType, nil),
			New[string, int]("beta", options.NoType, nil),
			New[string, int]("gamma", options.NoType, nil),
		}
		var nodeIDs []string
		for _, n := range nodes {
			nodeIDs = append(nodeIDs, n.ID())
		}

		root.Add(nodes...)
		defer root.Remove(nodeIDs...)
		m := root.adjacency()

		if len(*m) != len(nodes) {
			t.Errorf("unexpected graph length -- wanted %v ; got %v", len(nodes), len(*m))
		}
		for k := range *m {
			var ok bool
			for _, id := range nodeIDs {
				if id == k.ID() {
					ok = true
					break
				}
			}
			if !ok {
				t.Errorf("unable to find node %s in the adjacency list %v", k.ID(), nodeIDs)
			}
		}
	})
	t.Run("ThreeNodesOneEdge", func(t *testing.T) {
		nodes := []model.Graph[string, int]{
			New[string, int]("alpha", options.NoType, nil),
			New[string, int]("beta", options.NoType, nil),
			New[string, int]("gamma", options.NoType, nil),
		}
		var nodeIDs []string
		for _, n := range nodes {
			nodeIDs = append(nodeIDs, n.ID())
		}
		var edges = map[string][]string{
			"alpha": {"beta"},
		}

		root.Add(nodes...)
		for from, tos := range edges {
			for _, to := range tos {
				root.Connect(from, to, 1)
			}
		}
		defer root.Remove(nodeIDs...)

		m := root.adjacency()

		if len(*m) != len(nodes) {
			t.Errorf("unexpected graph length -- wanted %v ; got %v", len(nodes), len(*m))
		}
		for k := range *m {
			var ok bool
			for _, id := range nodeIDs {
				if id == k.ID() {
					ok = true
					break
				}
			}
			if !ok {
				t.Errorf("unable to find node %s in the adjacency list %v", k.ID(), nodeIDs != nil)
			}

			nodeEdges, err := root.Edges(k.ID())
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			setEdges := edges[k.ID()]
			for _, e := range nodeEdges {
				var ok bool
				for _, id := range setEdges {
					if id == e.ID() {
						ok = true
						break
					}
				}
				if !ok {
					t.Errorf("unable to find edge from %s to %s in the adjacency list %v", k.ID(), e.ID(), nodeIDs != nil)
				}
			}
		}
	})

	t.Run("ThreeNodesFourEdges", func(t *testing.T) {
		nodes := []model.Graph[string, int]{
			New[string, int]("alpha", options.NoType, nil),
			New[string, int]("beta", options.NoType, nil),
			New[string, int]("gamma", options.NoType, nil),
		}
		var nodeIDs []string
		for _, n := range nodes {
			nodeIDs = append(nodeIDs, n.ID())
		}
		var edges = map[string][]string{
			"alpha": {"beta", "gamma"},
			"beta":  {"gamma"},
			"gamma": {"alpha"},
		}

		root.Add(nodes...)
		for from, tos := range edges {
			for _, to := range tos {
				root.Connect(from, to, 1)
			}
		}
		defer root.Remove(nodeIDs...)

		m := root.adjacency()

		if len(*m) != len(nodes) {
			t.Errorf("unexpected graph length -- wanted %v ; got %v", len(nodes), len(*m))
		}
		for k := range *m {
			var ok bool
			for _, id := range nodeIDs {
				if id == k.ID() {
					ok = true
					break
				}
			}
			if !ok {
				t.Errorf("unable to find node %s in the adjacency list %v", k.ID(), nodeIDs != nil)
			}

			nodeEdges, err := root.Edges(k.ID())
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			setEdges := edges[k.ID()]
			for _, e := range nodeEdges {
				var ok bool
				for _, id := range setEdges {
					if id == e.ID() {
						ok = true
						break
					}
				}
				if !ok {
					t.Errorf("unable to find edge from %s to %s in the adjacency list %v", k.ID(), e.ID(), nodeIDs != nil)
				}
			}
		}
	})
}

func TestID(t *testing.T) {
	t.Run("StringID", func(t *testing.T) {
		g := New[string, int](testIDString, options.NoType, nil)

		if g.ID() != testIDString {
			t.Errorf("unexpected ID value: wanted %s ; got %s", testIDString, g.ID())
		}
	})
	t.Run("IntID", func(t *testing.T) {
		g := New[int, int](testIDInt, options.NoType, nil)

		if g.ID() != testIDInt {
			t.Errorf("unexpected ID value: wanted %v ; got %v", testIDInt, g.ID())
		}
	})
	t.Run("FloatID", func(t *testing.T) {
		g := New[float32, int](testIDFloat, options.NoType, nil)

		if g.ID() != testIDFloat {
			t.Errorf("unexpected ID value: wanted %v ; got %v", testIDFloat, g.ID())
		}
	})
}

func TestValue(t *testing.T) {
	t.Run("StringValue", func(t *testing.T) {
		g := New[int, int](0, testIDString, nil)

		v, ok := g.Value().(string)
		if !ok {
			t.Errorf("unexpected value type: wanted %T ; got %T", testIDString, g.Value())
		}

		if v != testIDString {
			t.Errorf("unexpected value: wanted %s ; got %s", testIDString, v)
		}
	})
	t.Run("IntValue", func(t *testing.T) {
		g := New[int, int](0, testIDInt, nil)

		v, ok := g.Value().(int)
		if !ok {
			t.Errorf("unexpected value type: wanted %T ; got %T", testIDInt, g.Value())
		}

		if v != testIDInt {
			t.Errorf("unexpected value: wanted %v ; got %v", testIDInt, v)
		}
	})
	t.Run("FloatValue", func(t *testing.T) {
		g := New[int, int](0, testIDFloat, nil)

		v, ok := g.Value().(float32)
		if !ok {
			t.Errorf("unexpected value type: wanted %T ; got %T", testIDFloat, g.Value())
		}

		if v != testIDFloat {
			t.Errorf("unexpected value: wanted %v ; got %v", testIDFloat, v)
		}
	})
	t.Run("StructValue", func(t *testing.T) {
		type custom struct {
			id       int
			name     string
			isActive bool
		}

		c := custom{
			id:       5,
			name:     "yes",
			isActive: true,
		}

		g := New[int, int](0, c, nil)

		v, ok := g.Value().(custom)
		if !ok {
			t.Errorf("unexpected value type: wanted %T ; got %T", testIDFloat, g.Value())
		}

		if !reflect.DeepEqual(c, v) {
			t.Errorf("unexpected value: wanted %v ; got %v", c, v)
		}
	})
	t.Run("AnonStructValue", func(t *testing.T) {
		g := New[int, int](0, struct {
			id       int
			name     string
			isActive bool
		}{
			id:       5,
			name:     "yes",
			isActive: true,
		}, nil)

		v, ok := g.Value().(struct {
			id       int
			name     string
			isActive bool
		})
		if !ok {
			t.Errorf("unexpected value type: wanted %T ; got %T", testIDFloat, g.Value())
		}

		wants := struct {
			id       int
			name     string
			isActive bool
		}{
			id:       5,
			name:     "yes",
			isActive: true,
		}
		if !reflect.DeepEqual(wants, v) {
			t.Errorf("unexpected value: wanted %v ; got %v", wants, v)
		}
	})
	t.Run("StructPointerValue", func(t *testing.T) {
		type custom struct {
			id       int
			name     string
			isActive bool
		}

		c := &custom{
			id:       5,
			name:     "yes",
			isActive: true,
		}

		g := New[int, int](0, c, nil)

		v, ok := g.Value().(*custom)
		if !ok {
			t.Errorf("unexpected value type: wanted %T ; got %T", testIDFloat, g.Value())
		}

		if !reflect.DeepEqual(*c, *v) {
			t.Errorf("unexpected value: wanted %v ; got %v", *c, *v)
		}
	})
	t.Run("InterfaceValue", func(t *testing.T) {
		c := &testObject{
			id:       5,
			name:     "yes",
			isActive: true,
		}

		i := ider(c)

		g := New[int, int](0, i, nil)

		itf, ok := g.Value().(ider)
		if !ok {
			t.Errorf("unexpected value type: wanted %T ; got %T", testIDFloat, g.Value())
		}

		v, ok := itf.(*testObject)
		if !ok {
			t.Errorf("unexpected value type: wanted %T ; got %T", testIDFloat, g.Value())
		}

		if !reflect.DeepEqual(*c, *v) {
			t.Errorf("unexpected value: wanted %v ; got %v", *c, *v)
		}
	})
}

func TestAdd(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("NestOneGraph", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, nil)
			node := New[string, int]("alpha", options.NoType, nil)

			err := root.Add(node)
			if err != nil {
				t.Errorf("unexpected error adding node %s: %v", node.ID(), err)
			}
			t.Run("VerifyNodeParent", func(t *testing.T) {
				g := node.Parent()
				if !reflect.DeepEqual(root, g) {
					t.Errorf("output mismatch error: wanted %v ; got %v", root, g)
				}
			})
			t.Run("VerifyRootNodes", func(t *testing.T) {
				nodes, err := root.List()
				if err != nil {
					t.Errorf("unexpected error listing nodes from %s: %v", root.ID(), err)
				}
				if len(nodes) != 1 {
					t.Errorf("unexpected length of nodes in graph %s: wanted %v ; got %v", root.ID(), 1, len(nodes))
				}
				if !reflect.DeepEqual(node, nodes[0]) {
					t.Errorf("output mismatch error: wanted %v ; got %v", node, nodes[0])
				}
			})
		})
		t.Run("NestThreeGraphs", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, nil)
			nodeA := New[string, int]("a", options.NoType, nil)
			nodeB := New[string, int]("b", options.NoType, nil)
			nodeC := New[string, int]("c", options.NoType, nil)
			t.Run("AddFirstToRoot", func(t *testing.T) {
				err := root.Add(nodeA)
				if err != nil {
					t.Errorf("unexpected error adding node %s to %s: %v", nodeA.ID(), root.ID(), err)
				}
				t.Run("VerifyNodeParent", func(t *testing.T) {
					g := nodeA.Parent()
					if !reflect.DeepEqual(root, g) {
						t.Errorf("output mismatch error: wanted %v ; got %v", root, g)
					}
				})
				t.Run("VerifyRootNodes", func(t *testing.T) {
					nodes, err := root.List()
					if err != nil {
						t.Errorf("unexpected error listing nodes from %s: %v", root.ID(), err)
					}
					if len(nodes) != 1 {
						t.Errorf("unexpected length of nodes in graph %s: wanted %v ; got %v", root.ID(), 1, len(nodes))
					}
					if !reflect.DeepEqual(nodeA, nodes[0]) {
						t.Errorf("output mismatch error: wanted %v ; got %v", nodeA, nodes[0])
					}
				})
			})

			t.Run("AddSecondToFirst", func(t *testing.T) {
				err := nodeA.Add(nodeB)
				if err != nil {
					t.Errorf("unexpected error adding node %s to %s: %v", nodeB.ID(), nodeA.ID(), err)
				}
				t.Run("VerifyNodeParent", func(t *testing.T) {
					gA := nodeB.Parent()
					if !reflect.DeepEqual(nodeA, gA) {
						t.Errorf("output mismatch error: wanted %v ; got %v", nodeA, gA)
					}
					gR := gA.Parent()
					if !reflect.DeepEqual(root, gR) {
						t.Errorf("output mismatch error: wanted %v ; got %v", root, gR)
					}
				})
				t.Run("VerifyRootNodes", func(t *testing.T) {
					nodesR, err := root.List()
					if err != nil {
						t.Errorf("unexpected error listing nodes from %s: %v", root.ID(), err)
					}
					if len(nodesR) != 1 {
						t.Errorf("unexpected length of nodes in graph %s: wanted %v ; got %v", root.ID(), 1, len(nodesR))
					}
					if !reflect.DeepEqual(nodeA, nodesR[0]) {
						t.Errorf("output mismatch error: wanted %v ; got %v", nodeA, nodesR[0])
					}
					nodesA, err := nodesR[0].List()
					if err != nil {
						t.Errorf("unexpected error listing nodes from %s: %v", nodesR[0].ID(), err)
					}
					if len(nodesA) != 1 {
						t.Errorf("unexpected length of nodes in graph %s: wanted %v ; got %v", nodesR[0].ID(), 1, len(nodesA))
					}
					if !reflect.DeepEqual(nodeB, nodesA[0]) {
						t.Errorf("output mismatch error: wanted %v ; got %v", nodeB, nodesA[0])
					}
				})
			})

			t.Run("AddThirdToSecond", func(t *testing.T) {
				err := nodeB.Add(nodeC)
				if err != nil {
					t.Errorf("unexpected error adding node %s to %s: %v", nodeC.ID(), nodeB.ID(), err)
				}
				t.Run("VerifyNodeParent", func(t *testing.T) {
					gB := nodeC.Parent()
					if !reflect.DeepEqual(nodeB, gB) {
						t.Errorf("output mismatch error: wanted %v ; got %v", nodeB, gB)
					}
					gA := gB.Parent()
					if !reflect.DeepEqual(nodeA, gA) {
						t.Errorf("output mismatch error: wanted %v ; got %v", nodeA, gA)
					}
					gR := gA.Parent()
					if !reflect.DeepEqual(root, gR) {
						t.Errorf("output mismatch error: wanted %v ; got %v", root, gR)
					}
				})
				t.Run("VerifyRootNodes", func(t *testing.T) {
					nodesR, err := root.List()
					if err != nil {
						t.Errorf("unexpected error listing nodes from %s: %v", root.ID(), err)
					}
					if len(nodesR) != 1 {
						t.Errorf("unexpected length of nodes in graph %s: wanted %v ; got %v", root.ID(), 1, len(nodesR))
					}
					if !reflect.DeepEqual(nodeA, nodesR[0]) {
						t.Errorf("output mismatch error: wanted %v ; got %v", nodeA, nodesR[0])
					}
					nodesA, err := nodesR[0].List()
					if err != nil {
						t.Errorf("unexpected error listing nodes from %s: %v", nodesR[0].ID(), err)
					}
					if len(nodesA) != 1 {
						t.Errorf("unexpected length of nodes in graph %s: wanted %v ; got %v", nodesR[0].ID(), 1, len(nodesA))
					}
					if !reflect.DeepEqual(nodeB, nodesA[0]) {
						t.Errorf("output mismatch error: wanted %v ; got %v", nodeB, nodesA[0])
					}
					nodesB, err := nodesA[0].List()
					if err != nil {
						t.Errorf("unexpected error listing nodes from %s: %v", nodesA[0].ID(), err)
					}
					if len(nodesB) != 1 {
						t.Errorf("unexpected length of nodes in graph %s: wanted %v ; got %v", nodesA[0].ID(), 1, len(nodesB))
					}
					if !reflect.DeepEqual(nodeC, nodesB[0]) {
						t.Errorf("output mismatch error: wanted %v ; got %v", nodeC, nodesB[0])
					}
				})
			})
		})
	})

	t.Run("Fail", func(t *testing.T) {
		t.Run("AddingSelfAsNode", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, nil)

			err := root.Add(root)
			if err == nil {
				t.Errorf("error expected when adding self as a node")
			}
			if !errors.Is(err, errs.InvalidOperation) {
				t.Errorf("unexpected error returned; wanted %v ; got %v", err, errs.InvalidOperation)
			}
		})
		t.Run("AddingToALockedGraph", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, nil)
			nodeA := New[string, int]("a", options.NoType, options.ReadOnly)
			nodeB := New[string, int]("b", options.NoType, nil)

			err := root.Add(nodeA)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			err = nodeA.Add(nodeB)
			if err == nil {
				t.Errorf("error expected when adding to a read-only graph")
			}
			if !errors.Is(err, errs.ReadOnly) {
				t.Errorf("unexpected error returned; wanted %v ; got %v", err, errs.ReadOnly)
			}
		})
	})
}
