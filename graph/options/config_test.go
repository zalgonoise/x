package options

import (
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	defaultConf := &GraphConfig{}

	t.Run("Empty", func(t *testing.T) {
		conf := New()

		if conf == nil {
			t.Errorf("output config cannot be nil")
		}
		if !reflect.DeepEqual(defaultConf, conf) {
			t.Errorf("mismatch error -- wanted %v ; got %v", *defaultConf, *conf)
		}
	})
	t.Run("NilParam", func(t *testing.T) {
		conf := New(nil)

		if conf == nil {
			t.Errorf("output config cannot be nil")
		}
		if !reflect.DeepEqual(defaultConf, conf) {
			t.Errorf("mismatch error -- wanted %v ; got %v", *defaultConf, *conf)
		}
	})
	t.Run("MultipleNilParams", func(t *testing.T) {
		conf := New(nil, nil, nil, nil)

		if conf == nil {
			t.Errorf("output config cannot be nil")
		}
		if !reflect.DeepEqual(defaultConf, conf) {
			t.Errorf("mismatch error -- wanted %v ; got %v", *defaultConf, *conf)
		}
	})
	t.Run("WithPreset", func(t *testing.T) {
		wants := &GraphConfig{
			GraphType: GraphList,
		}
		conf := New(CfgAdjacencyList)

		if conf == nil {
			t.Errorf("output config cannot be nil")
		}
		if !reflect.DeepEqual(wants, conf) {
			t.Errorf("mismatch error -- wanted %v ; got %v", *defaultConf, *conf)
		}
	})
	t.Run("WithPresetAndOverride", func(t *testing.T) {
		wants := &GraphConfig{
			GraphType:         GraphMatrix,
			NoCrossGraphEdges: true,
		}
		conf := New(CfgAdjacencyList, GraphMatrix)

		if conf == nil {
			t.Errorf("output config cannot be nil")
		}
		if !reflect.DeepEqual(wants, conf) {
			t.Errorf("mismatch error -- wanted %v ; got %v", *wants, *conf)
		}
	})
}

func TestGraphConfigApply(t *testing.T) {
	defaultConf := &GraphConfig{}

	t.Run("EmptyToEmpty", func(t *testing.T) {
		var conf = new(GraphConfig)
		defaultConf.Apply(conf)

		if !reflect.DeepEqual(defaultConf, conf) {
			t.Errorf("mismatch error -- wanted %v ; got %v", *defaultConf, *conf)
		}
	})

	t.Run("EmptyToPreset", func(t *testing.T) {
		wants := &GraphConfig{
			GraphType: GraphList,
		}
		var conf = new(GraphConfig)
		CfgAdjacencyList.Apply(conf)

		if !reflect.DeepEqual(wants, conf) {
			t.Errorf("mismatch error -- wanted %v ; got %v", *defaultConf, *conf)
		}
	})

	t.Run("UpdateWithPreset", func(t *testing.T) {
		wants := &GraphConfig{
			GraphType: GraphList,
			MaxNodes:  2,
		}
		var conf = &GraphConfig{
			MaxNodes: 2,
		}
		CfgAdjacencyList.Apply(conf)

		if !reflect.DeepEqual(wants, conf) {
			t.Errorf("mismatch error -- wanted %v ; got %v", *defaultConf, *conf)
		}
	})

	t.Run("OverwriteWithPreset", func(t *testing.T) {
		wants := &GraphConfig{
			GraphType: GraphList,
		}
		var conf = &GraphConfig{
			GraphType:         GraphMatrix,
			NoCrossGraphEdges: true,
		}
		CfgAdjacencyList.Apply(conf)

		if !reflect.DeepEqual(wants, conf) {
			t.Errorf("mismatch error -- wanted %v ; got %v", *defaultConf, *conf)
		}
	})
}
