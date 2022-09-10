package options

import (
	"reflect"
	"testing"
)

func FuzzMaxNodes(f *testing.F) {
	f.Add(0)
	f.Add(3)
	f.Add(5)
	f.Add(50)
	f.Add(100)

	f.Fuzz(func(t *testing.T, a int) {
		s := MaxNodes(a)

		conf := New(s)

		if conf.MaxNodes != a {
			t.Errorf("output mismatch error -- wanted %v ; got %v", a, conf.MaxNodes)
		}
	})
}

func FuzzMaxDepth(f *testing.F) {
	f.Add(0)
	f.Add(3)
	f.Add(5)
	f.Add(50)
	f.Add(100)

	f.Fuzz(func(t *testing.T, a int) {
		s := MaxDepth(a)

		conf := New(s)

		if conf.MaxDepth != a {
			t.Errorf("output mismatch error -- wanted %v ; got %v", a, conf.MaxDepth)
		}
	})
}

func TestOptions(t *testing.T) {
	t.Run("MaxNodes", func(t *testing.T) {
		t.Run("NegativeInput", func(t *testing.T) {
			s := MaxNodes(-5)

			conf := New(s)
			if conf.MaxNodes != 0 {
				t.Errorf("output mismatch error -- wanted %v ; got %v", 0, conf.MaxNodes)
			}
		})
	})
	t.Run("MaxDepth", func(t *testing.T) {
		t.Run("NegativeInput", func(t *testing.T) {
			s := MaxDepth(-5)

			conf := New(s)
			if conf.MaxDepth != 0 {
				t.Errorf("output mismatch error -- wanted %v ; got %v", 0, conf.MaxDepth)
			}
		})
	})
	t.Run("Type", func(t *testing.T) {
		t.Run("List", func(t *testing.T) {
			wants := &GraphConfig{
				GraphType: GraphList,
			}

			conf := New(GraphList)
			if !reflect.DeepEqual(wants, conf) {
				t.Errorf("output mismatch error for config -- wanted %v ; got %v", wants, conf)
			}
		})
		t.Run("Matrix", func(t *testing.T) {
			wants := &GraphConfig{
				GraphType:         GraphMatrix,
				NoCrossGraphEdges: true,
			}

			conf := New(GraphMatrix)
			if !reflect.DeepEqual(wants, conf) {
				t.Errorf("output mismatch error for config -- wanted %v ; got %v", wants, conf)
			}
		})
	})
	t.Run("Direction", func(t *testing.T) {
		t.Run("Directed", func(t *testing.T) {
			wants := &GraphConfig{}

			conf := New(Directional)
			if !reflect.DeepEqual(wants, conf) {
				t.Errorf("output mismatch error for config -- wanted %v ; got %v", wants, conf)
			}
		})
		t.Run("Undirected", func(t *testing.T) {
			wants := &GraphConfig{
				IsNonDirectional: true,
			}

			conf := New(NonDirectional)
			if !reflect.DeepEqual(wants, conf) {
				t.Errorf("output mismatch error for config -- wanted %v ; got %v", wants, conf)
			}
		})
	})
	t.Run("Cycles", func(t *testing.T) {
		t.Run("Cyclical", func(t *testing.T) {
			wants := &GraphConfig{}

			conf := New(Cyclical)
			if !reflect.DeepEqual(wants, conf) {
				t.Errorf("output mismatch error for config -- wanted %v ; got %v", wants, conf)
			}
		})
		t.Run("NonCyclical", func(t *testing.T) {
			wants := &GraphConfig{
				IsNonCyclical: true,
			}

			conf := New(NonCyclical)
			if !reflect.DeepEqual(wants, conf) {
				t.Errorf("output mismatch error for config -- wanted %v ; got %v", wants, conf)
			}
		})
	})
	t.Run("Weights", func(t *testing.T) {
		t.Run("Weighted", func(t *testing.T) {
			wants := &GraphConfig{}

			conf := New(Weighted)
			if !reflect.DeepEqual(wants, conf) {
				t.Errorf("output mismatch error for config -- wanted %v ; got %v", wants, conf)
			}
		})
		t.Run("Unweighted", func(t *testing.T) {
			wants := &GraphConfig{
				IsUnweighted: true,
			}

			conf := New(Unweighted)
			if !reflect.DeepEqual(wants, conf) {
				t.Errorf("output mismatch error for config -- wanted %v ; got %v", wants, conf)
			}
		})
	})
	t.Run("Mutability", func(t *testing.T) {
		t.Run("Mutable", func(t *testing.T) {
			wants := &GraphConfig{}

			conf := New(Mutable)
			if !reflect.DeepEqual(wants, conf) {
				t.Errorf("output mismatch error for config -- wanted %v ; got %v", wants, conf)
			}
		})
		t.Run("Immutable", func(t *testing.T) {
			wants := &GraphConfig{
				Immutable: true,
			}

			conf := New(Immutable)
			if !reflect.DeepEqual(wants, conf) {
				t.Errorf("output mismatch error for config -- wanted %v ; got %v", wants, conf)
			}
		})
	})
	t.Run("WritePrivilege", func(t *testing.T) {
		t.Run("ReadWrite", func(t *testing.T) {
			wants := &GraphConfig{}

			conf := New(ReadWrite)
			if !reflect.DeepEqual(wants, conf) {
				t.Errorf("output mismatch error for config -- wanted %v ; got %v", wants, conf)
			}
		})
		t.Run("ReadOnly", func(t *testing.T) {
			wants := &GraphConfig{
				ReadOnly: true,
			}

			conf := New(ReadOnly)
			if !reflect.DeepEqual(wants, conf) {
				t.Errorf("output mismatch error for config -- wanted %v ; got %v", wants, conf)
			}
		})
	})
	t.Run("WeightAsDistance", func(t *testing.T) {
		t.Run("AsLabel", func(t *testing.T) {
			wants := &GraphConfig{}

			conf := New(LabelWeight)
			if !reflect.DeepEqual(wants, conf) {
				t.Errorf("output mismatch error for config -- wanted %v ; got %v", wants, conf)
			}
		})
		t.Run("AsDistance", func(t *testing.T) {
			wants := &GraphConfig{
				WeightAsDistance: true,
			}

			conf := New(DistanceWeight)
			if !reflect.DeepEqual(wants, conf) {
				t.Errorf("output mismatch error for config -- wanted %v ; got %v", wants, conf)
			}
		})
	})
}
