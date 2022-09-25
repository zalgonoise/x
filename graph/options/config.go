package options

import (
	"reflect"
)

type (
	// Setting describes a configurable option
	Setting interface {
		// Apply sets this option in the input GraphConfig pointer
		Apply(*GraphConfig)
	}

	// GraphConfig defines how a graph can be configured
	GraphConfig struct {
		GraphType         TypeSetting
		IsNonDirectional  bool
		IsNonCyclical     bool
		IsUnweighted      bool
		Immutable         bool
		ReadOnly          bool
		WeightAsDistance  bool
		NoCrossGraphEdges bool
		MaxNodes          int
		MaxDepth          int
		IDConstraint      reflect.Type
	}
)

// New function creates a new pointer to a GraphConfig from the list of
// input Setting provided
func New(s ...Setting) *GraphConfig {
	if len(s) == 0 {
		return &GraphConfig{}
	}

	conf := new(GraphConfig)

	input := MultiOption(s...)
	if input == nil {
		return &GraphConfig{}
	}

	input.Apply(conf)

	return conf
}

// Apply overrides a pointer to a GraphConfig with this GraphConfig's settings
func (c *GraphConfig) Apply(t *GraphConfig) {
	*t = *c
}
