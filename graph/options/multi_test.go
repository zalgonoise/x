package options

import (
	"reflect"
	"testing"
)

func TestMultiOption(t *testing.T) {
	// defaultConf := &GraphConfig{}

	t.Run("Success", func(t *testing.T) {
		t.Run("SingleParam", func(t *testing.T) {
			wants := &GraphConfig{
				GraphType:         GraphMatrix,
				NoCrossGraphEdges: true,
			}

			s := MultiOption(GraphMatrix)
			if s == nil {
				t.Errorf("output setting cannot be nil")
			}

			conf := New(s)
			if conf == nil {
				t.Errorf("output config cannot be nil")
			}
			if !reflect.DeepEqual(wants, conf) {
				t.Errorf("mismatch error -- wanted %v ; got %v", *wants, *conf)
			}
		})
		t.Run("MultipleParams", func(t *testing.T) {
			wants := &GraphConfig{
				GraphType:         GraphMatrix,
				NoCrossGraphEdges: true,
				MaxDepth:          2,
			}

			s := MultiOption(GraphMatrix, MaxDepth(2))
			if s == nil {
				t.Errorf("output setting cannot be nil")
			}

			conf := New(s)
			if conf == nil {
				t.Errorf("output config cannot be nil")
			}
			if !reflect.DeepEqual(wants, conf) {
				t.Errorf("mismatch error -- wanted %v ; got %v", *wants, *conf)
			}
		})
		t.Run("OverwrittingParams", func(t *testing.T) {
			wants := &GraphConfig{
				GraphType:         GraphMatrix,
				NoCrossGraphEdges: true,
			}

			s := MultiOption(GraphList, GraphMatrix)
			if s == nil {
				t.Errorf("output setting cannot be nil")
			}

			conf := New(s)
			if conf == nil {
				t.Errorf("output config cannot be nil")
			}
			if !reflect.DeepEqual(*wants, *conf) {
				t.Errorf("mismatch error -- wanted %v ; got %v", *wants, *conf)
			}
		})
		t.Run("NestedMultiOption", func(t *testing.T) {
			wants := &GraphConfig{
				GraphType:         GraphMatrix,
				NoCrossGraphEdges: true,
				Immutable:         true,
				MaxNodes:          2,
				MaxDepth:          2,
			}

			s1 := MultiOption(MaxNodes(2), MaxDepth(2))
			if s1 == nil {
				t.Errorf("output setting 1 cannot be nil")
			}
			s2 := MultiOption(GraphMatrix, Immutable, s1)
			if s2 == nil {
				t.Errorf("output setting 2 cannot be nil")
			}

			conf := New(s2)
			if conf == nil {
				t.Errorf("output config cannot be nil")
			}
			if !reflect.DeepEqual(*wants, *conf) {
				t.Errorf("mismatch error -- wanted %v ; got %v", *wants, *conf)
			}
		})

	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Empty", func(t *testing.T) {
			s := MultiOption()

			if s != nil {
				t.Errorf("output setting must be nil")
			}
		})

		t.Run("NilParam", func(t *testing.T) {
			s := MultiOption(nil)

			if s != nil {
				t.Errorf("output setting must be nil")
			}
		})

		t.Run("MultipleNilParams", func(t *testing.T) {
			s := MultiOption(nil, nil, nil, nil)

			if s != nil {
				t.Errorf("output setting must be nil")
			}
		})
	})

}
