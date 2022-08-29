package options

import (
	"reflect"
)

type (
	Setting interface {
		Apply(*GraphConfig)
	}

	GraphConfig struct {
		GraphType        TypeSetting
		IsNonDirectional bool
		IsNonCyclical    bool
		IsUnweighted     bool
		Immutable        bool
		ReadOnly         bool
		WeightAsDistance bool
		MaxNodes         int
		MaxDepth         int
		IDConstraint     reflect.Type
	}
)

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

func (c *GraphConfig) Apply(t *GraphConfig) {
	*t = *c
}
