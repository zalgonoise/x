package options

import (
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
			wants := GraphList
			s := wants

			conf := New(s)
			if conf.GraphType != wants {
				t.Errorf("output mismatch error for graph type -- wanted %v ; got %v", wants, conf.GraphType)
			}
		})
		t.Run("Matrix", func(t *testing.T) {
			wants := GraphMatrix
			s := wants

			conf := New(s)
			if conf.GraphType != wants {
				t.Errorf("output mismatch error for graph type -- wanted %v ; got %v", wants, conf.GraphType)
			}
		})

	})
	t.Run("Direction", func(t *testing.T) {
		t.Run("Directed", func(t *testing.T) {
			wants := Directional
			s := wants

			conf := New(s)
			if conf.IsNonDirectional {
				t.Errorf("output mismatch error for direction -- wanted %v ; got %v", wants, conf.IsNonDirectional)
			}
		})
		t.Run("Undirected", func(t *testing.T) {
			wants := NonDirectional
			s := wants

			conf := New(s)
			if !conf.IsNonDirectional {
				t.Errorf("output mismatch error for direction -- wanted %v ; got %v", wants, conf.IsNonDirectional)
			}
		})

	})
	t.Run("Cycles", func(t *testing.T) {
		t.Run("Cyclical", func(t *testing.T) {
			wants := Cyclical
			s := wants

			conf := New(s)
			if conf.IsNonCyclical {
				t.Errorf("output mismatch error for cycles -- wanted %v ; got %v", wants, conf.IsNonCyclical)
			}
		})
		t.Run("NonCyclical", func(t *testing.T) {
			wants := NonCyclical
			s := wants

			conf := New(s)
			if !conf.IsNonCyclical {
				t.Errorf("output mismatch error for cycles -- wanted %v ; got %v", wants, conf.IsNonCyclical)
			}
		})

	})
	t.Run("Weights", func(t *testing.T) {
		t.Run("Weighted", func(t *testing.T) {
			wants := Weighted
			s := wants

			conf := New(s)
			if conf.IsUnweighted {
				t.Errorf("output mismatch error for weighted edges -- wanted %v ; got %v", wants, conf.IsUnweighted)
			}
		})
		t.Run("Unweighted", func(t *testing.T) {
			wants := Unweighted
			s := wants

			conf := New(s)
			if !conf.IsUnweighted {
				t.Errorf("output mismatch error for weighted edges -- wanted %v ; got %v", wants, conf.IsUnweighted)
			}
		})

	})
	t.Run("Mutability", func(t *testing.T) {
		t.Run("Mutable", func(t *testing.T) {
			wants := Mutable
			s := wants

			conf := New(s)
			if conf.Immutable {
				t.Errorf("output mismatch error for mutability -- wanted %v ; got %v", wants, conf.Immutable)
			}
		})
		t.Run("Immutable", func(t *testing.T) {
			wants := Immutable
			s := wants

			conf := New(s)
			if !conf.Immutable {
				t.Errorf("output mismatch error for mutability -- wanted %v ; got %v", wants, conf.Immutable)
			}
		})

	})
	t.Run("WritePrivilege", func(t *testing.T) {
		t.Run("ReadWrite", func(t *testing.T) {
			wants := ReadWrite
			s := wants

			conf := New(s)
			if conf.ReadOnly {
				t.Errorf("output mismatch error for write privilege -- wanted %v ; got %v", wants, conf.ReadOnly)
			}
		})
		t.Run("ReadOnly", func(t *testing.T) {
			wants := ReadOnly
			s := wants

			conf := New(s)
			if !conf.ReadOnly {
				t.Errorf("output mismatch error for write privilege -- wanted %v ; got %v", wants, conf.ReadOnly)
			}
		})

	})
	t.Run("WeightAsDistance", func(t *testing.T) {
		t.Run("AsDistance", func(t *testing.T) {
			wants := DistanceWeight
			s := wants

			conf := New(s)
			if !conf.WeightAsDistance {
				t.Errorf("output mismatch error for weight label -- wanted %v ; got %v", wants, conf.WeightAsDistance)
			}
		})
		t.Run("AsLabel", func(t *testing.T) {
			wants := LabelWeight
			s := wants

			conf := New(s)
			if conf.ReadOnly {
				t.Errorf("output mismatch error for weight label -- wanted %v ; got %v", wants, conf.WeightAsDistance)
			}
		})

	})
}
