package list

import (
	"testing"

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
