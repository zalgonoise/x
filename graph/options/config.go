package options

import "reflect"

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

func New(s ...Setting) (*GraphConfig, error) {
	if len(s) == 0 {
		return &GraphConfig{}, nil
	}

	conf := new(GraphConfig)

	input := MultiOption(s...)

	input.Apply(conf)

	_, err := conf.Validate()
	if err != nil {
		return nil, err
	}

	return conf, nil
}

func (c *GraphConfig) Apply(t *GraphConfig) {
	*t = *c
}
