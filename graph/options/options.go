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
		IDConstraint     reflect.Type
	}

	TypeSetting      int
	DirectionSetting int
	CycleSetting     int
	WeightedEdges    int
	Mutability       int
	WritePrivilege   int
	IDConstraint     struct {
		v reflect.Type
	}
)

const (
	UnsetType   TypeSetting = iota
	GraphMatrix TypeSetting = iota
	GraphList   TypeSetting = iota
	GraphNode   TypeSetting = iota
	GraphHub    TypeSetting = iota
)
const (
	Directional    DirectionSetting = iota
	NonDirectional DirectionSetting = iota
)
const (
	Cyclical    CycleSetting = iota
	NonCyclical CycleSetting = iota
)
const (
	Weighted   WeightedEdges = iota
	Unweighted WeightedEdges = iota
)
const (
	Mutable   Mutability = iota
	Immutable Mutability = iota
)
const (
	ReadWrite WritePrivilege = iota
	ReadOnly  WritePrivilege = iota
)

func IDType(v any) Setting {
	return &IDConstraint{v: reflect.TypeOf(v)}
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

func (s TypeSetting) Apply(c *GraphConfig) {
	switch s {
	case UnsetType:
		c.GraphType = GraphMatrix
	default:
		c.GraphType = s
	}
}

func (s DirectionSetting) Apply(c *GraphConfig) {
	if s == NonDirectional {
		c.IsNonDirectional = true
	}
}

func (s CycleSetting) Apply(c *GraphConfig) {
	if s == NonCyclical {
		c.IsNonCyclical = true
	}
}

func (s WeightedEdges) Apply(c *GraphConfig) {
	if s == Unweighted {
		c.IsUnweighted = true
	}
}

func (s *IDConstraint) Apply(c *GraphConfig) {
	if s.v == nil {
		return
	}
	c.IDConstraint = s.v
}

func (s Mutability) Apply(c *GraphConfig) {
	if s == Immutable {
		c.Immutable = true
	}
}

func (s WritePrivilege) Apply(c *GraphConfig) {
	if s == ReadOnly {
		c.ReadOnly = true
	}
}
