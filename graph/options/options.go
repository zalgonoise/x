package options

type Setting interface {
	Apply(*GraphConfig)
}

type GraphConfig struct {
	GraphType        TypeSetting
	IsNonDirectional bool
	IsNonCyclical    bool
	IsUnweighted     bool
	Value            any
	IDConstraint     any
}

func New(s ...Setting) (*GraphConfig, error) {
	conf := new(GraphConfig)

	input := MultiOption(s...)

	input.Apply(conf)

	_, err := conf.Validate()
	if err != nil {
		return nil, err
	}

	return conf, nil
}

type TypeSetting int

const (
	UnsetType   TypeSetting = iota
	GraphMatrix TypeSetting = iota
	GraphList   TypeSetting = iota
	GraphNode   TypeSetting = iota
	GraphHub    TypeSetting = iota
)

func (s TypeSetting) Apply(c *GraphConfig) {
	switch c.GraphType {
	case UnsetType:
		c.GraphType = GraphMatrix
	default:
		c.GraphType = s
	}
}

type DirectionSetting int

const (
	Directional    DirectionSetting = iota
	NonDirectional DirectionSetting = iota
)

func (s DirectionSetting) Apply(c *GraphConfig) {
	if s == NonDirectional {
		c.IsNonDirectional = true
	}
}

type CycleSetting int

const (
	Cyclical    CycleSetting = iota
	NonCyclical CycleSetting = iota
)

func (s CycleSetting) Apply(c *GraphConfig) {
	if s == NonCyclical {
		c.IsNonCyclical = true
	}
}

type WeightedEdges int

const (
	Weighted   WeightedEdges = iota
	Unweighted WeightedEdges = iota
)

func (s WeightedEdges) Apply(c *GraphConfig) {
	if s == Unweighted {
		c.IsUnweighted = true
	}
}

type ValueGraph struct {
	v any
}

func (s *ValueGraph) Apply(c *GraphConfig) {
	if s.v == nil {
		return
	}
	c.Value = s.v
}

func WithValue(v any) Setting {
	return &ValueGraph{v: v}
}

type IDConstraint struct {
	v any
}

func (s *IDConstraint) Apply(c *GraphConfig) {
	if s.v == nil {
		return
	}
	c.Value = s.v
}

func IDType(v any) Setting {
	return &IDConstraint{v: v}
}
