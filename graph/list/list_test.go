package list

import (
	"testing"

	"github.com/zalgonoise/x/graph/model"
	"github.com/zalgonoise/x/graph/options"
)

const (
	testIDString string  = "test-id"
	testIDInt    int     = 0
	testIDFloat  float32 = 0
)

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
