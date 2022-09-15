package matrix

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

func verify(node model.Graph[string, int], wants ...model.Graph[string, int]) bool {
	for _, n := range wants {
		if reflect.DeepEqual(n, node) {
			return true
		}
	}
	return false
}

func verifyBatch(nodes []model.Graph[string, int], wants ...model.Graph[string, int]) bool {
	oks := make([]bool, len(wants))
	for idx, n := range nodes {
		oks[idx] = verify(n, wants...)
	}

	for _, ok := range oks {
		if !ok {
			return false
		}
	}
	return true
}

func TestNew(t *testing.T) {
	t.Run("StringID", func(t *testing.T) {
		t.Run("IntWeight", func(t *testing.T) {
			t.Run("Success", func(t *testing.T) {
				g := New[string, int](testIDString, testIDString, options.GraphList)

				_, ok := g.(*mapGraph[string, int])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *mapGraph[string, int], got %T", g)
				}
			})
			t.Run("SuccessWithNilValue", func(t *testing.T) {
				g := New[string, int](testIDString, nil, options.GraphList)

				_, ok := g.(*mapGraph[string, int])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *mapGraph[string, int], got %T", g)
				}
			})
			t.Run("SuccessWithEmptyValue", func(t *testing.T) {
				g := New[string, int](testIDString, options.NoType, options.GraphList)

				_, ok := g.(*mapGraph[string, int])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *mapGraph[string, int], got %T", g)
				}
			})
			t.Run("SuccessWithInvalidOption", func(t *testing.T) {
				g := New[string, int](testIDString, options.NoType, options.GraphMatrix)

				_, ok := g.(*mapGraph[string, int])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *mapGraph[string, int], got %T", g)
				}
			})
			t.Run("SuccessWithNoConfig", func(t *testing.T) {
				g := New[string, int](testIDString, options.NoType, nil)

				_, ok := g.(*mapGraph[string, int])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *mapGraph[string, int], got %T", g)
				}
			})
		})
		t.Run("Float64Weight", func(t *testing.T) {
			t.Run("Success", func(t *testing.T) {
				g := New[string, float64](testIDString, testIDString, options.GraphList)

				_, ok := g.(*mapGraph[string, float64])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *mapGraph[string, int], got %T", g)
				}
			})
			t.Run("SuccessWithNilValue", func(t *testing.T) {
				g := New[string, float64](testIDString, nil, options.GraphList)

				_, ok := g.(*mapGraph[string, float64])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *mapGraph[string, int], got %T", g)
				}
			})
			t.Run("SuccessWithEmptyValue", func(t *testing.T) {
				g := New[string, float64](testIDString, options.NoType, options.GraphList)

				_, ok := g.(*mapGraph[string, float64])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *mapGraph[string, int], got %T", g)
				}
			})
		})

	})
	t.Run("IntID", func(t *testing.T) {
		t.Run("IntWeight", func(t *testing.T) {
			t.Run("Success", func(t *testing.T) {
				g := New[int, int](testIDInt, testIDInt, options.GraphList)

				_, ok := g.(*mapGraph[int, int])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *mapGraph[int, int], got %T", g)
				}
			})
			t.Run("SuccessWithNilValue", func(t *testing.T) {
				g := New[int, int](testIDInt, nil, options.GraphList)

				_, ok := g.(*mapGraph[int, int])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *mapGraph[int, int], got %T", g)
				}
			})
			t.Run("SuccessWithEmptyValue", func(t *testing.T) {
				g := New[int, int](testIDInt, options.NoType, options.GraphList)

				_, ok := g.(*mapGraph[int, int])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *mapGraph[int, int], got %T", g)
				}
			})
		})
		t.Run("Float64Weight", func(t *testing.T) {
			t.Run("Success", func(t *testing.T) {
				g := New[int, float64](testIDInt, testIDInt, options.GraphList)

				_, ok := g.(*mapGraph[int, float64])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *mapGraph[int, int], got %T", g)
				}
			})
			t.Run("SuccessWithNilValue", func(t *testing.T) {
				g := New[int, float64](testIDInt, nil, options.GraphList)

				_, ok := g.(*mapGraph[int, float64])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *mapGraph[int, int], got %T", g)
				}
			})
			t.Run("SuccessWithEmptyValue", func(t *testing.T) {
				g := New[int, float64](testIDInt, options.NoType, options.GraphList)

				_, ok := g.(*mapGraph[int, float64])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *mapGraph[int, int], got %T", g)
				}
			})
		})
	})
	t.Run("FloatID", func(t *testing.T) {
		t.Run("IntWeight", func(t *testing.T) {
			t.Run("Success", func(t *testing.T) {
				g := New[float32, int](testIDFloat, testIDFloat, options.GraphList)

				_, ok := g.(*mapGraph[float32, int])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *mapGraph[float32, int], got %T", g)
				}
			})
			t.Run("SuccessWithNilValue", func(t *testing.T) {
				g := New[float32, int](testIDFloat, nil, options.GraphList)

				_, ok := g.(*mapGraph[float32, int])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *mapGraph[float32, int], got %T", g)
				}
			})
			t.Run("SuccessWithEmptyValue", func(t *testing.T) {
				g := New[float32, int](testIDFloat, options.NoType, options.GraphList)

				_, ok := g.(*mapGraph[float32, int])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *mapGraph[float32, int], got %T", g)
				}
			})
		})
		t.Run("Float64Weight", func(t *testing.T) {
			t.Run("Success", func(t *testing.T) {
				g := New[float32, float64](testIDFloat, testIDFloat, options.GraphList)

				_, ok := g.(*mapGraph[float32, float64])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *mapGraph[float32, int], got %T", g)
				}
			})
			t.Run("SuccessWithNilValue", func(t *testing.T) {
				g := New[float32, float64](testIDFloat, nil, options.GraphList)

				_, ok := g.(*mapGraph[float32, float64])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *mapGraph[float32, int], got %T", g)
				}
			})
			t.Run("SuccessWithEmptyValue", func(t *testing.T) {
				g := New[float32, float64](testIDFloat, options.NoType, options.GraphList)

				_, ok := g.(*mapGraph[float32, float64])
				if !ok {
					t.Errorf("unexpected graph type -- wanted *mapGraph[float32, int], got %T", g)
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

		err := root.Add(nodes...)
		if err != nil {
			t.Errorf("unexpected error adding nodes: %v", err)
		}

		defer func() { _ = root.Remove(nodeIDs...) }()
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

		err := root.Add(nodes...)
		if err != nil {
			t.Errorf("unexpected error adding nodes: %v", err)
		}
		defer func() { _ = root.Remove(nodeIDs...) }()
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

		err := root.Add(nodes...)
		if err != nil {
			t.Errorf("unexpected error adding nodes: %v", err)
		}
		for from, tos := range edges {
			for _, to := range tos {
				err = root.Connect(from, to, 1)
				if err != nil {
					t.Errorf("unexpected error joining nodes %s to %s: %v", from, to, err)
				}
			}
		}
		defer func() { _ = root.Remove(nodeIDs...) }()

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

		err := root.Add(nodes...)
		if err != nil {
			t.Errorf("unexpected error adding nodes: %v", err)
		}
		for from, tos := range edges {
			for _, to := range tos {
				err = root.Connect(from, to, 1)
				if err != nil {
					t.Errorf("unexpected error connecting nodes %s to %s: %v", from, to, err)
				}
			}
		}
		defer func() { _ = root.Remove(nodeIDs...) }()

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

func TestConfig(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		wants := options.New(options.CfgAdjacencyMatrix)

		g := New[string, int](testIDString, options.NoType, nil)
		conf := g.Config()

		if !reflect.DeepEqual(wants, conf) {
			t.Errorf("output mismatch error: wanted %v ; got %v", wants, conf)
		}
	})

	t.Run("WithConfig", func(t *testing.T) {
		wants := options.New(options.CfgAdjacencyMatrix, options.ReadOnly)

		g := New[string, int](testIDString, options.NoType, options.ReadOnly)
		conf := g.Config()

		if !reflect.DeepEqual(wants, conf) {
			t.Errorf("output mismatch error: wanted %v ; got %v", wants, conf)
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
		t.Run("AddingNodesWithLimit", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, options.MaxNodes(1))
			nodeA := New[string, int]("a", options.NoType, nil)
			nodeB := New[string, int]("b", options.NoType, nil)

			err := root.Add(nodeA)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			err = root.Add(nodeB)
			if err == nil {
				t.Errorf("unexpected an error when adding the second node")
			}
			if !errors.Is(err, errs.MaxNodesReached) {
				t.Errorf("unexpected error: wanted %v ; got %v", errs.MaxNodesReached, err)
			}
		})
		t.Run("AddingWithDepthLimit", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, options.MaxDepth(1))
			nodeA := New[string, int]("a", options.NoType, nil)
			nodeB := New[string, int]("b", options.NoType, nil)

			err := root.Add(nodeA)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			err = nodeA.Add(nodeB)
			if err == nil {
				t.Errorf("unexpected an error when adding the second node")
			}
			if !errors.Is(err, errs.MaxDepthReached) {
				t.Errorf("unexpected error: wanted %v ; got %v", errs.MaxDepthReached, err)
			}
		})
		t.Run("NodeAlreadyExists", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, options.MaxDepth(1))
			nodeA := New[string, int]("a", options.NoType, nil)

			err := root.Add(nodeA)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			err = root.Add(nodeA)
			if err == nil {
				t.Errorf("unexpected an error when adding the second node")
			}
			if !errors.Is(err, errs.AlreadyExists) {
				t.Errorf("unexpected error: wanted %v ; got %v", errs.AlreadyExists, err)
			}
		})
	})
}

func TestRemove(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("RemoveOneOfThree", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, nil)
			nodeA := New[string, int]("a", options.NoType, nil)
			nodeB := New[string, int]("b", options.NoType, nil)
			nodeC := New[string, int]("c", options.NoType, nil)

			err := root.Add(nodeA, nodeB, nodeC)
			if err != nil {
				t.Errorf("unexpected error adding nodes %s %s and %s: %v", nodeA.ID(), nodeB.ID(), nodeC.ID(), err)
			}
			err = root.Remove(nodeB.ID())
			if err != nil {
				t.Errorf("unexpected error removing node %s: %v", nodeB.ID(), err)
			}
			t.Run("VerifyGraphLength", func(t *testing.T) {
				nodes, err := root.List()
				if err != nil {
					t.Errorf("unexpected error listing nodes from %s: %v", root.ID(), err)
				}
				if len(nodes) != 2 {
					t.Errorf("unexpected graph nodes length: wanted %v ; got %v", 2, len(nodes))
				}
			})
		})
		t.Run("RemoveThreeOfThree", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, nil)
			nodeA := New[string, int]("a", options.NoType, nil)
			nodeB := New[string, int]("b", options.NoType, nil)
			nodeC := New[string, int]("c", options.NoType, nil)

			err := root.Add(nodeA, nodeB, nodeC)
			if err != nil {
				t.Errorf("unexpected error adding nodes %s %s and %s: %v", nodeA.ID(), nodeB.ID(), nodeC.ID(), err)
			}
			err = root.Remove(nodeA.ID(), nodeB.ID(), nodeC.ID())
			if err != nil {
				t.Errorf("unexpected error removing nodes %s %s and %s: %v", nodeA.ID(), nodeB.ID(), nodeC.ID(), err)
			}
			t.Run("VerifyGraphLength", func(t *testing.T) {
				nodes, err := root.List()
				if err != nil {
					t.Errorf("unexpected error listing nodes from %s: %v", root.ID(), err)
				}
				if len(nodes) != 0 {
					t.Errorf("unexpected graph nodes length: wanted %v ; got %v", 0, len(nodes))
				}
			})
		})
		t.Run("RemovingANodeWithEdges", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, nil)
			nodeA := New[string, int]("a", options.NoType, nil)
			nodeB := New[string, int]("b", options.NoType, nil)

			err := root.Add(nodeA, nodeB)
			if err != nil {
				t.Errorf("unexpected error adding nodes: %v", err)
			}
			err = root.Connect(nodeB.ID(), nodeA.ID(), 1)
			if err != nil {
				t.Errorf("unexpected error connecting edges: %v", err)
			}

			err = root.Remove(nodeA.ID())
			if err != nil {
				t.Errorf("unexpected error when removing a node with edges")
			}
		})
	})

	t.Run("Fail", func(t *testing.T) {
		t.Run("RemovingFromALockedGraph", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, nil)
			nodeA := New[string, int]("a", options.NoType, options.ReadOnly)
			nodeB := New[string, int]("b", options.NoType, nil)

			err := nodeA.Add(nodeB)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			err = root.Add(nodeA)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			err = nodeA.Remove(nodeB.ID())
			if err == nil {
				t.Errorf("error expected when removing from a read-only graph")
			}
			if !errors.Is(err, errs.ReadOnly) {
				t.Errorf("unexpected error returned; wanted %v ; got %v", err, errs.ReadOnly)
			}
		})
		t.Run("RemovingFromAnImmutableGraph", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, options.Immutable)
			nodeA := New[string, int]("a", options.NoType, nil)

			err := root.Add(nodeA)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			err = root.Remove(nodeA.ID())
			if err == nil {
				t.Errorf("error expected when removing from an immutable graph")
			}
			if !errors.Is(err, errs.Immutable) {
				t.Errorf("unexpected error returned; wanted %v ; got %v", err, errs.Immutable)
			}
		})
		t.Run("RemovingFromANonExistingNode", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, nil)
			nodeA := New[string, int]("a", options.NoType, nil)
			nodeB := New[string, int]("b", options.NoType, nil)

			err := root.Add(nodeA)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			err = root.Remove(nodeB.ID())
			if err == nil {
				t.Errorf("error expected when removing a node not belonging to the graph")
			}
			if !errors.Is(err, errs.DoesNotExist) {
				t.Errorf("unexpected error returned; wanted %v ; got %v", err, errs.DoesNotExist)
			}
		})
	})
}

func TestGet(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("GetOneOfOne", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, nil)
			nodeA := New[string, int]("a", options.NoType, nil)

			err := root.Add(nodeA)
			if err != nil {
				t.Errorf("unexpected error adding nodes %s: %v", nodeA.ID(), err)
			}
			node, err := root.Get(nodeA.ID())
			if err != nil {
				t.Errorf("unexpected error getting node %s: %v", nodeA.ID(), err)
			}
			if !reflect.DeepEqual(node, nodeA) {
				t.Errorf("output mismatch error: wanted %v ; got %v", nodeA, node)
			}
		})

		t.Run("GetOneOfThree", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, nil)
			nodeA := New[string, int]("a", options.NoType, nil)
			nodeB := New[string, int]("a", options.NoType, nil)
			nodeC := New[string, int]("a", options.NoType, nil)

			err := root.Add(nodeA, nodeB, nodeC)
			if err != nil {
				t.Errorf("unexpected error adding node %s %s and %s: %v", nodeA.ID(), nodeB.ID(), nodeC.ID(), err)
			}
			node, err := root.Get(nodeB.ID())
			if err != nil {
				t.Errorf("unexpected error getting node %s: %v", nodeB.ID(), err)
			}
			if !reflect.DeepEqual(node, nodeB) {
				t.Errorf("output mismatch error: wanted %v ; got %v", nodeB, node)
			}
		})
	})

	t.Run("Fail", func(t *testing.T) {
		t.Run("GettingANonExistingNode", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, nil)
			nodeA := New[string, int]("a", options.NoType, nil)
			nodeB := New[string, int]("b", options.NoType, nil)

			err := root.Add(nodeA)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			node, err := root.Get(nodeB.ID())
			if err == nil {
				t.Errorf("error expected when removing a node not belonging to the graph")
			}
			if node != nil {
				t.Errorf("expected returned node to be nil; is %v", node)
			}
			if !errors.Is(err, errs.DoesNotExist) {
				t.Errorf("unexpected error returned; wanted %v ; got %v", err, errs.DoesNotExist)
			}
		})
	})
}

func TestList(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("ListOne", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, nil)
			nodeA := New[string, int]("a", options.NoType, nil)

			err := root.Add(nodeA)
			if err != nil {
				t.Errorf("unexpected error adding nodes %s: %v", nodeA.ID(), err)
			}
			nodes, err := root.List()
			if err != nil {
				t.Errorf("unexpected error getting node %s: %v", nodeA.ID(), err)
			}
			if len(nodes) != 1 {
				t.Errorf("unexpected nodes list length: wanted %v ; got %v", 1, len(nodes))
			}
			if !reflect.DeepEqual(nodes[0], nodeA) {
				t.Errorf("output mismatch error: wanted %v ; got %v", nodeA, nodes[0])
			}
		})

		t.Run("ListThree", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, nil)
			nodeA := New[string, int]("a", options.NoType, nil)
			nodeB := New[string, int]("b", options.NoType, nil)
			nodeC := New[string, int]("c", options.NoType, nil)

			err := root.Add(nodeA, nodeB, nodeC)
			if err != nil {
				t.Errorf("unexpected error adding node %s %s and %s: %v", nodeA.ID(), nodeB.ID(), nodeC.ID(), err)
			}
			nodes, err := root.List()
			if err != nil {
				t.Errorf("unexpected error getting node %s: %v", nodeB.ID(), err)
			}
			if len(nodes) != 3 {
				t.Errorf("unexpected nodes list length: wanted %v ; got %v", 1, len(nodes))
			}

			if !verifyBatch(nodes, nodeA, nodeB, nodeC) {
				t.Errorf("failed to match retrieved nodes to input nodes")
			}
		})
	})
}

func TestConnect(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("ConnectOne", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, nil)
			nodeA := New[string, int]("a", options.NoType, nil)
			nodeB := New[string, int]("b", options.NoType, nil)
			nodeC := New[string, int]("c", options.NoType, nil)

			err := root.Add(nodeA, nodeB, nodeC)
			if err != nil {
				t.Errorf("unexpected error adding node %s %s and %s: %v", nodeA.ID(), nodeB.ID(), nodeC.ID(), err)
			}

			err = root.Connect(nodeA.ID(), nodeB.ID(), 2)
			if err != nil {
				t.Errorf("unexpected error linking nodes %s to %s with weight %v: %v", nodeA.ID(), nodeB.ID(), 2, err)
			}

			e, err := root.Edges(nodeA.ID())
			if err != nil {
				t.Errorf("unexpected error getting edges from node %s: %v", nodeA.ID(), err)
			}

			if len(e) != 1 {
				t.Errorf("unexpected edge length: wanted %v ; got %v", 1, len(e))
			}

			if nodeB.ID() != e[0].ID() {
				t.Errorf("output mismatch error: wanted %s ; got %s", nodeB.ID(), e[0].ID())
			}

			w, err := root.Weight(nodeA.ID(), nodeB.ID())
			if err != nil {
				t.Errorf("unexpected error getting weight from node link %s to %s: %v", nodeA.ID(), nodeB.ID(), err)
			}

			if w != 2 {
				t.Errorf("unexpected weight from connected edges: wanted %v ; got %v", 2, w)
			}
		})
		t.Run("ConnectBackAndForth", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, nil)
			nodeA := New[string, int]("a", options.NoType, nil)
			nodeB := New[string, int]("b", options.NoType, nil)
			nodeC := New[string, int]("c", options.NoType, nil)

			err := root.Add(nodeA, nodeB, nodeC)
			if err != nil {
				t.Errorf("unexpected error adding node %s %s and %s: %v", nodeA.ID(), nodeB.ID(), nodeC.ID(), err)
			}

			err = root.Connect(nodeA.ID(), nodeB.ID(), 2)
			if err != nil {
				t.Errorf("unexpected error linking nodes %s to %s with weight %v: %v", nodeA.ID(), nodeB.ID(), 2, err)
			}

			err = root.Connect(nodeB.ID(), nodeA.ID(), 3)
			if err != nil {
				t.Errorf("unexpected error linking nodes %s to %s with weight %v: %v", nodeB.ID(), nodeA.ID(), 3, err)
			}

			e, err := root.Edges(nodeA.ID())
			if err != nil {
				t.Errorf("unexpected error getting edges from node %s: %v", nodeA.ID(), err)
			}

			if len(e) != 1 {
				t.Errorf("unexpected edge length: wanted %v ; got %v", 1, len(e))
			}

			if nodeB.ID() != e[0].ID() {
				t.Errorf("output mismatch error: wanted %s ; got %s", nodeB.ID(), e[0].ID())
			}

			w, err := root.Weight(nodeA.ID(), nodeB.ID())
			if err != nil {
				t.Errorf("unexpected error getting weight from node link %s to %s: %v", nodeA.ID(), nodeB.ID(), err)
			}

			if w != 2 {
				t.Errorf("unexpected weight from connected edges: wanted %v ; got %v", 2, w)
			}

			e, err = root.Edges(nodeB.ID())
			if err != nil {
				t.Errorf("unexpected error getting edges from node %s: %v", nodeB.ID(), err)
			}

			if len(e) != 1 {
				t.Errorf("unexpected edge length: wanted %v ; got %v", 1, len(e))
			}

			if nodeA.ID() != e[0].ID() {
				t.Errorf("output mismatch error: wanted %s ; got %s", nodeA.ID(), e[0].ID())
			}

			w, err = root.Weight(nodeB.ID(), nodeA.ID())
			if err != nil {
				t.Errorf("unexpected error getting weight from node link %s to %s: %v", nodeB.ID(), nodeA.ID(), err)
			}

			if w != 3 {
				t.Errorf("unexpected weight from connected edges: wanted %v ; got %v", 3, w)
			}
		})
		t.Run("AddNonDirectional", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, options.NonDirectional)
			nodeA := New[string, int]("a", options.NoType, nil)
			nodeB := New[string, int]("b", options.NoType, nil)

			err := root.Add(nodeA, nodeB)
			if err != nil {
				t.Errorf("unexpected error adding nodes %v and %v to %v: %v", nodeA.ID(), nodeB.ID(), root.ID(), err)
			}
			err = root.Connect(nodeA.ID(), nodeB.ID(), 1)
			if err != nil {
				t.Errorf("unexpected error connecting nodes %v and %v: %v", nodeA.ID(), nodeB.ID(), err)
			}
			e, err := root.Edges(nodeA.ID())
			if err != nil {
				t.Errorf("unexpected error getting edges from node %s: %v", nodeA.ID(), err)
			}
			if len(e) != 1 {
				t.Errorf("unexpected number of edges in node %v: wanted %v ; got %v", nodeA.ID(), 1, len(e))
			}
			if !reflect.DeepEqual(nodeB, e[0]) {
				t.Errorf("output mismatch error: wanted %v ; got %v", nodeB.ID(), e[0].ID())
			}
			e, err = root.Edges(nodeB.ID())
			if err != nil {
				t.Errorf("unexpected error getting edges from node %s: %v", nodeB.ID(), err)
			}
			if len(e) != 1 {
				t.Errorf("unexpected number of edges in node %v: wanted %v ; got %v", nodeB.ID(), 1, len(e))
			}
			if !reflect.DeepEqual(nodeA, e[0]) {
				t.Errorf("output mismatch error: wanted %v ; got %v", nodeA.ID(), e[0].ID())
			}
		})
	})
	t.Run("Fail", func(t *testing.T) {
		t.Run("ConnectingInALockedGraph", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, nil)
			nodeA := New[string, int]("a", options.NoType, options.ReadOnly)
			nodeB := New[string, int]("b", options.NoType, nil)
			nodeC := New[string, int]("c", options.NoType, nil)

			err := nodeA.Add(nodeB, nodeC)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			err = root.Add(nodeA)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			err = nodeA.Connect(nodeB.ID(), nodeC.ID(), 1)
			if err == nil {
				t.Errorf("expected an error when connecting nodes in a read-only graph")
			}
			if !errors.Is(err, errs.ReadOnly) {
				t.Errorf("unexpected error returned; wanted %v ; got %v", err, errs.ReadOnly)
			}
		})
		t.Run("ConnectingFromANonExistantNode", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, nil)
			nodeA := New[string, int]("a", options.NoType, nil)

			err := root.Add(nodeA)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			err = root.Connect("b", nodeA.ID(), 1)
			if err == nil {
				t.Errorf("expected an error when connecting nodes in a read-only graph")
			}
			if !errors.Is(err, errs.DoesNotExist) {
				t.Errorf("unexpected error returned; wanted %v ; got %v", err, errs.ReadOnly)
			}
		})
		t.Run("ConnectingToANonExistantNode", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, nil)
			nodeA := New[string, int]("a", options.NoType, nil)

			err := root.Add(nodeA)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			err = root.Connect(nodeA.ID(), "b", 1)
			if err == nil {
				t.Errorf("expected an error when connecting nodes in a read-only graph")
			}
			if !errors.Is(err, errs.DoesNotExist) {
				t.Errorf("unexpected error returned; wanted %v ; got %v", err, errs.ReadOnly)
			}
		})
		t.Run("CyclesInANonCyclicalGraph", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, options.NonCyclical)
			nodeA := New[string, int]("a", options.NoType, nil)
			nodeB := New[string, int]("b", options.NoType, nil)
			nodeC := New[string, int]("c", options.NoType, nil)

			err := root.Add(nodeA, nodeB, nodeC)
			if err != nil {
				t.Errorf("unexpected error adding nodes %v %v and %v to %v: %v", nodeA.ID(), nodeB.ID(), nodeC.ID(), root.ID(), err)
			}
			err = root.Connect(nodeA.ID(), nodeB.ID(), 1)
			if err != nil {
				t.Errorf("unexpected error connecting nodes %v to %v: %v", nodeA.ID(), nodeB.ID(), err)
			}
			err = root.Connect(nodeB.ID(), nodeC.ID(), 1)
			if err != nil {
				t.Errorf("unexpected error connecting nodes %v to %v: %v", nodeA.ID(), nodeB.ID(), err)
			}
			err = root.Connect(nodeC.ID(), nodeA.ID(), 1)
			if err == nil {
				t.Errorf("error expected when connecting nodes %v to %v, for finding a cycle in the edges", nodeC.ID(), nodeA.ID())
			}
			if !errors.Is(err, errs.CyclicalEdge) {
				t.Errorf("unexpected output error: wanted %v ; got %v", errs.CyclicalEdge, err)
			}
		})
		t.Run("ConnectAcrossGraphs", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, nil)
			hubA := New[string, int]("hub-a", options.NoType, nil)
			hubB := New[string, int]("hub-b", options.NoType, nil)
			hubC := New[string, int]("hub-c", options.NoType, nil)

			subhubA1 := New[string, int]("sub-hub-a1", options.NoType, nil)
			subhubA2 := New[string, int]("sub-hub-a2", options.NoType, nil)
			subhubB1 := New[string, int]("sub-hub-b1", options.NoType, nil)
			subhubB2 := New[string, int]("sub-hub-b2", options.NoType, nil)
			subhubC1 := New[string, int]("sub-hub-c1", options.NoType, nil)
			subhubC2 := New[string, int]("sub-hub-c2", options.NoType, nil)

			nodeA11 := New[string, int]("a11", options.NoType, nil)
			nodeA12 := New[string, int]("a12", options.NoType, nil)
			nodeA21 := New[string, int]("a21", options.NoType, nil)
			nodeA22 := New[string, int]("a22", options.NoType, nil)
			nodeB11 := New[string, int]("b11", options.NoType, nil)
			nodeB12 := New[string, int]("b12", options.NoType, nil)
			nodeB21 := New[string, int]("b21", options.NoType, nil)
			nodeB22 := New[string, int]("b22", options.NoType, nil)
			nodeC11 := New[string, int]("c11", options.NoType, nil)
			nodeC12 := New[string, int]("c12", options.NoType, nil)
			nodeC21 := New[string, int]("c21", options.NoType, nil)
			nodeC22 := New[string, int]("c22", options.NoType, nil)

			err := root.Add(hubA, hubB, hubC)
			if err != nil {
				t.Errorf("unexpected error adding node %s %s and %s to %s: %v", hubA.ID(), hubB.ID(), hubC.ID(), root.ID(), err)
			}

			err = hubA.Add(subhubA1, subhubA2)
			if err != nil {
				t.Errorf("unexpected error adding node %s and %s to %s: %v", subhubA1.ID(), subhubA2.ID(), hubA.ID(), err)
			}
			err = hubB.Add(subhubB1, subhubB2)
			if err != nil {
				t.Errorf("unexpected error adding node %s and %s to %s: %v", subhubB1.ID(), subhubB2.ID(), hubB.ID(), err)
			}
			err = hubC.Add(subhubC1, subhubC2)
			if err != nil {
				t.Errorf("unexpected error adding node %s and %s to %s: %v", subhubC1.ID(), subhubC2.ID(), hubC.ID(), err)
			}

			err = subhubA1.Add(nodeA11, nodeA12)
			if err != nil {
				t.Errorf("unexpected error adding node %s and %s to %s: %v", nodeA11.ID(), nodeA12.ID(), subhubA1.ID(), err)
			}
			err = subhubA2.Add(nodeA21, nodeA22)
			if err != nil {
				t.Errorf("unexpected error adding node %s and %s to %s: %v", nodeA21.ID(), nodeA22.ID(), subhubA2.ID(), err)
			}
			err = subhubB1.Add(nodeB11, nodeB12)
			if err != nil {
				t.Errorf("unexpected error adding node %s and %s to %s: %v", nodeB11.ID(), nodeB12.ID(), subhubB1.ID(), err)
			}
			err = subhubB2.Add(nodeB21, nodeB22)
			if err != nil {
				t.Errorf("unexpected error adding node %s and %s to %s: %v", nodeB21.ID(), nodeB22.ID(), subhubB2.ID(), err)
			}
			err = subhubC1.Add(nodeC11, nodeC12)
			if err != nil {
				t.Errorf("unexpected error adding node %s and %s to %s: %v", nodeC11.ID(), nodeC12.ID(), subhubC1.ID(), err)
			}
			err = subhubC2.Add(nodeC21, nodeC22)
			if err != nil {
				t.Errorf("unexpected error adding node %s and %s to %s: %v", nodeC21.ID(), nodeC22.ID(), subhubC2.ID(), err)
			}

			err = subhubA1.Connect(nodeA11.ID(), nodeC22.ID(), 2)
			if err == nil {
				t.Errorf("error expected when connecting nodes %v and %v: %v", nodeA11.ID(), nodeC22.ID(), err)
			}

			if !errors.Is(err, errs.DoesNotExist) {
				t.Errorf("unexpected error: wanted %v ; got %v", errs.DoesNotExist, err)
			}
		})
	})
}

func TestDisconnect(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("DisconnectOne", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, nil)
			nodeA := New[string, int]("a", options.NoType, nil)
			nodeB := New[string, int]("b", options.NoType, nil)
			nodeC := New[string, int]("c", options.NoType, nil)

			err := root.Add(nodeA, nodeB, nodeC)
			if err != nil {
				t.Errorf("unexpected error adding node %s %s and %s: %v", nodeA.ID(), nodeB.ID(), nodeC.ID(), err)
			}

			err = root.Connect(nodeA.ID(), nodeB.ID(), 2)
			if err != nil {
				t.Errorf("unexpected error linking nodes %s to %s with weight %v: %v", nodeA.ID(), nodeB.ID(), 2, err)
			}

			err = root.Disconnect(nodeA.ID(), nodeB.ID())
			if err != nil {
				t.Errorf("unexpected error unlinking nodes %s to %s: %v", nodeA.ID(), nodeB.ID(), err)
			}

			e, err := root.Edges(nodeA.ID())
			if err != nil {
				t.Errorf("unexpected error getting edges from node %s: %v", nodeA.ID(), err)
			}

			if len(e) != 0 {
				t.Errorf("unexpected edge length: wanted %v ; got %v -- %v %v", 0, len(e), e[0], e[1])
			}
		})
		t.Run("DisconnectTwoOfThree", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, nil)
			nodeA := New[string, int]("a", options.NoType, nil)
			nodeB := New[string, int]("b", options.NoType, nil)
			nodeC := New[string, int]("c", options.NoType, nil)
			nodeD := New[string, int]("d", options.NoType, nil)

			err := root.Add(nodeA, nodeB, nodeC, nodeD)
			if err != nil {
				t.Errorf("unexpected error adding node %s %s %s and %s: %v", nodeA.ID(), nodeB.ID(), nodeC.ID(), nodeD.ID(), err)
			}

			err = root.Connect(nodeA.ID(), nodeB.ID(), 2)
			if err != nil {
				t.Errorf("unexpected error linking nodes %s to %s with weight %v: %v", nodeA.ID(), nodeB.ID(), 2, err)
			}
			err = root.Connect(nodeA.ID(), nodeC.ID(), 3)
			if err != nil {
				t.Errorf("unexpected error linking nodes %s to %s with weight %v: %v", nodeA.ID(), nodeC.ID(), 3, err)
			}
			err = root.Connect(nodeA.ID(), nodeD.ID(), 4)
			if err != nil {
				t.Errorf("unexpected error linking nodes %s to %s with weight %v: %v", nodeA.ID(), nodeD.ID(), 4, err)
			}

			err = root.Disconnect(nodeA.ID(), nodeC.ID())
			if err != nil {
				t.Errorf("unexpected error unlinking nodes %s to %s: %v", nodeA.ID(), nodeC.ID(), err)
			}
			err = root.Disconnect(nodeA.ID(), nodeB.ID())
			if err != nil {
				t.Errorf("unexpected error unlinking nodes %s to %s: %v", nodeA.ID(), nodeB.ID(), err)
			}

			e, err := root.Edges(nodeA.ID())
			if err != nil {
				t.Errorf("unexpected error getting edges from node %s: %v", nodeA.ID(), err)
			}

			if len(e) != 1 {
				t.Errorf("unexpected edge length: wanted %v ; got %v", 1, len(e))
			}

			if nodeD.ID() != e[0].ID() {
				t.Errorf("output mismatch error: wanted %s ; got %s", nodeD.ID(), e[0].ID())
			}

			w, err := root.Weight(nodeA.ID(), nodeD.ID())
			if err != nil {
				t.Errorf("unexpected error getting weight from node link %s to %s: %v", nodeA.ID(), nodeD.ID(), err)
			}

			if w != 4 {
				t.Errorf("unexpected weight from connected edges: wanted %v ; got %v", 4, w)
			}
		})
		t.Run("DisconnectNonDirectional", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, options.NonDirectional)
			nodeA := New[string, int]("a", options.NoType, nil)
			nodeB := New[string, int]("b", options.NoType, nil)

			err := root.Add(nodeA, nodeB)
			if err != nil {
				t.Errorf("unexpected error adding nodes %v and %v to %v: %v", nodeA.ID(), nodeB.ID(), root.ID(), err)
			}
			err = root.Connect(nodeA.ID(), nodeB.ID(), 1)
			if err != nil {
				t.Errorf("unexpected error connecting nodes %v and %v: %v", nodeA.ID(), nodeB.ID(), err)
			}
			err = root.Disconnect(nodeA.ID(), nodeB.ID())
			if err != nil {
				t.Errorf("unexpected error disconnecting nodes %v and %v: %v", nodeA.ID(), nodeB.ID(), err)
			}
			e, err := root.Edges(nodeA.ID())
			if err != nil {
				t.Errorf("unexpected error getting edges from node %s: %v", nodeA.ID(), err)
			}
			if len(e) != 0 {
				t.Errorf("unexpected number of edges in node %v: wanted %v ; got %v", nodeA.ID(), 0, len(e))
			}
			e, err = root.Edges(nodeB.ID())
			if err != nil {
				t.Errorf("unexpected error getting edges from node %s: %v", nodeB.ID(), err)
			}
			if len(e) != 0 {
				t.Errorf("unexpected number of edges in node %v: wanted %v ; got %v", nodeB.ID(), 0, len(e))
			}
		})
	})
	t.Run("Fail", func(t *testing.T) {
		t.Run("DisconnectingFromALockedGraph", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, nil)
			nodeA := New[string, int]("a", options.NoType, options.ReadOnly)
			nodeB := New[string, int]("b", options.NoType, nil)
			nodeC := New[string, int]("c", options.NoType, nil)

			err := nodeA.Add(nodeB, nodeC)
			if err != nil {
				t.Errorf("unexpected error adding nodes %v and %v to %v: %v", nodeB.ID(), nodeC.ID(), nodeA.ID(), err)
			}
			err = nodeA.Connect(nodeB.ID(), nodeC.ID(), 1)
			if err != nil {
				t.Errorf("unexpected error connecting nodes %v to %v: %v", nodeB.ID(), nodeC.ID(), err)
			}
			err = root.Add(nodeA)
			if err != nil {
				t.Errorf("unexpected error adding node %v to %v: %v", nodeA.ID(), root.ID(), err)
			}
			err = nodeA.Disconnect(nodeB.ID(), nodeC.ID())
			if err == nil {
				t.Errorf("error expected when disconnecting nodes in a read-only graph")
			}
			if !errors.Is(err, errs.ReadOnly) {
				t.Errorf("unexpected error returned; wanted %v ; got %v", err, errs.ReadOnly)
			}
		})
		t.Run("DisconnectingFromAnImmutableGraph", func(t *testing.T) {
			nodeA := New[string, int]("a", options.NoType, options.Immutable)
			nodeB := New[string, int]("b", options.NoType, nil)
			nodeC := New[string, int]("c", options.NoType, nil)

			err := nodeA.Add(nodeB, nodeC)
			if err != nil {
				t.Errorf("unexpected error adding nodes %v and %v to %v: %v", nodeB.ID(), nodeC.ID(), nodeA.ID(), err)
			}
			err = nodeA.Connect(nodeB.ID(), nodeC.ID(), 1)
			if err != nil {
				t.Errorf("unexpected error connecting nodes %v to %v: %v", nodeB.ID(), nodeC.ID(), err)
			}
			err = nodeA.Disconnect(nodeB.ID(), nodeC.ID())
			if err == nil {
				t.Errorf("error expected when disconnecting nodes in an immutable graph")
			}
			if !errors.Is(err, errs.Immutable) {
				t.Errorf("unexpected error returned; wanted %v ; got %v", err, errs.Immutable)
			}
		})
		t.Run("DisconnectAcrossGraphs", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, nil)
			hubA := New[string, int]("hub-a", options.NoType, nil)
			hubB := New[string, int]("hub-b", options.NoType, nil)
			hubC := New[string, int]("hub-c", options.NoType, nil)

			subhubA1 := New[string, int]("sub-hub-a1", options.NoType, nil)
			subhubA2 := New[string, int]("sub-hub-a2", options.NoType, nil)
			subhubB1 := New[string, int]("sub-hub-b1", options.NoType, nil)
			subhubB2 := New[string, int]("sub-hub-b2", options.NoType, nil)
			subhubC1 := New[string, int]("sub-hub-c1", options.NoType, nil)
			subhubC2 := New[string, int]("sub-hub-c2", options.NoType, nil)

			nodeA11 := New[string, int]("a11", options.NoType, nil)
			nodeA12 := New[string, int]("a12", options.NoType, nil)
			nodeA21 := New[string, int]("a21", options.NoType, nil)
			nodeA22 := New[string, int]("a22", options.NoType, nil)
			nodeB11 := New[string, int]("b11", options.NoType, nil)
			nodeB12 := New[string, int]("b12", options.NoType, nil)
			nodeB21 := New[string, int]("b21", options.NoType, nil)
			nodeB22 := New[string, int]("b22", options.NoType, nil)
			nodeC11 := New[string, int]("c11", options.NoType, nil)
			nodeC12 := New[string, int]("c12", options.NoType, nil)
			nodeC21 := New[string, int]("c21", options.NoType, nil)
			nodeC22 := New[string, int]("c22", options.NoType, nil)

			err := root.Add(hubA, hubB, hubC)
			if err != nil {
				t.Errorf("unexpected error adding node %s %s and %s to %s: %v", hubA.ID(), hubB.ID(), hubC.ID(), root.ID(), err)
			}

			err = hubA.Add(subhubA1, subhubA2)
			if err != nil {
				t.Errorf("unexpected error adding node %s and %s to %s: %v", subhubA1.ID(), subhubA2.ID(), hubA.ID(), err)
			}
			err = hubB.Add(subhubB1, subhubB2)
			if err != nil {
				t.Errorf("unexpected error adding node %s and %s to %s: %v", subhubB1.ID(), subhubB2.ID(), hubB.ID(), err)
			}
			err = hubC.Add(subhubC1, subhubC2)
			if err != nil {
				t.Errorf("unexpected error adding node %s and %s to %s: %v", subhubC1.ID(), subhubC2.ID(), hubC.ID(), err)
			}

			err = subhubA1.Add(nodeA11, nodeA12)
			if err != nil {
				t.Errorf("unexpected error adding node %s and %s to %s: %v", nodeA11.ID(), nodeA12.ID(), subhubA1.ID(), err)
			}
			err = subhubA2.Add(nodeA21, nodeA22)
			if err != nil {
				t.Errorf("unexpected error adding node %s and %s to %s: %v", nodeA21.ID(), nodeA22.ID(), subhubA2.ID(), err)
			}
			err = subhubB1.Add(nodeB11, nodeB12)
			if err != nil {
				t.Errorf("unexpected error adding node %s and %s to %s: %v", nodeB11.ID(), nodeB12.ID(), subhubB1.ID(), err)
			}
			err = subhubB2.Add(nodeB21, nodeB22)
			if err != nil {
				t.Errorf("unexpected error adding node %s and %s to %s: %v", nodeB21.ID(), nodeB22.ID(), subhubB2.ID(), err)
			}
			err = subhubC1.Add(nodeC11, nodeC12)
			if err != nil {
				t.Errorf("unexpected error adding node %s and %s to %s: %v", nodeC11.ID(), nodeC12.ID(), subhubC1.ID(), err)
			}
			err = subhubC2.Add(nodeC21, nodeC22)
			if err != nil {
				t.Errorf("unexpected error adding node %s and %s to %s: %v", nodeC21.ID(), nodeC22.ID(), subhubC2.ID(), err)
			}

			err = subhubA1.Disconnect(nodeA11.ID(), nodeC22.ID())
			if err == nil {
				t.Errorf("error expected when disconnecting nodes %v and %v: %v", nodeA11.ID(), nodeC22.ID(), err)
			}

			if !errors.Is(err, errs.DoesNotExist) {
				t.Errorf("unexpected error: wanted %v ; got %v", errs.DoesNotExist, err)
			}

		})
	})
}

func TestEdges(t *testing.T) {
	t.Run("Fail", func(t *testing.T) {
		t.Run("FromNodeDoesNotExist", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, nil)
			nodeA := New[string, int]("a", options.NoType, nil)

			err := root.Add(nodeA)
			if err != nil {
				t.Errorf("unexpected error adding nodes %v to %v: %v", nodeA.ID(), root.ID(), err)
			}
			err = root.Connect("b", nodeA.ID(), 1)
			if err == nil {
				t.Errorf("error expected when connecting a node which doesn't exist to another node")
			}
			if !errors.Is(err, errs.DoesNotExist) {
				t.Errorf("unexpected error returned; wanted %v ; got %v", err, errs.DoesNotExist)
			}
		})
		t.Run("ToNodeDoesNotExist", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, nil)
			nodeA := New[string, int]("a", options.NoType, nil)

			err := root.Add(nodeA)
			if err != nil {
				t.Errorf("unexpected error adding nodes %v to %v: %v", nodeA.ID(), root.ID(), err)
			}
			err = root.Connect(nodeA.ID(), "b", 1)
			if err == nil {
				t.Errorf("error expected when connecting a node to one which doesn't exist")
			}
			if !errors.Is(err, errs.DoesNotExist) {
				t.Errorf("unexpected error returned; wanted %v ; got %v", err, errs.DoesNotExist)
			}
		})
	})
}

func TestWeight(t *testing.T) {
	t.Run("Fail", func(t *testing.T) {
		t.Run("FromNodeDoesNotExist", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, nil)
			nodeA := New[string, int]("a", options.NoType, nil)

			err := root.Add(nodeA)
			if err != nil {
				t.Errorf("unexpected error adding nodes %v to %v: %v", nodeA.ID(), root.ID(), err)
			}
			_, err = root.Weight("b", nodeA.ID())
			if err == nil {
				t.Errorf("error expected when getting the weight of a connection where the to node doesn't exist")
			}
			if !errors.Is(err, errs.DoesNotExist) {
				t.Errorf("unexpected error returned; wanted %v ; got %v", errs.DoesNotExist, err)
			}
		})
		t.Run("ToNodeDoesNotExist", func(t *testing.T) {
			root := New[string, int](testIDString, options.NoType, nil)
			nodeA := New[string, int]("a", options.NoType, nil)

			err := root.Add(nodeA)
			if err != nil {
				t.Errorf("unexpected error adding nodes %v to %v: %v", nodeA.ID(), root.ID(), err)
			}
			_, err = root.Weight(nodeA.ID(), "b")
			if err == nil {
				t.Errorf("error expected when getting the weight of a connection where the to node doesn't exist")
			}
			if !errors.Is(err, errs.DoesNotExist) {
				t.Errorf("unexpected error returned; wanted %v ; got %v", errs.DoesNotExist, err)
			}
		})
	})
}
