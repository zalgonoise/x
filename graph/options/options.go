package options

import (
	"reflect"
)

type (
	TypeSetting      int
	DirectionSetting int
	CycleSetting     int
	WeightedEdges    int
	Mutability       int
	WritePrivilege   int
	NodeLimit        int
	DepthLimit       int
	IDConstraint     struct {
		v reflect.Type
	}
)

const (
	GraphMatrix TypeSetting = iota
	GraphList
)
const (
	Directional DirectionSetting = iota
	NonDirectional
)
const (
	Cyclical CycleSetting = iota
	NonCyclical
)
const (
	Weighted WeightedEdges = iota
	Unweighted
)
const (
	Mutable Mutability = iota
	Immutable
)
const (
	ReadWrite WritePrivilege = iota
	ReadOnly
)

func MaxNodes(v int) Setting {
	s := NodeLimit(v)
	return &s
}

func MaxDepth(v int) Setting {
	s := DepthLimit(v)
	return &s
}

func IDType(v any) Setting {
	return &IDConstraint{v: reflect.TypeOf(v)}
}

func (s TypeSetting) Apply(c *GraphConfig) {
	c.GraphType = s
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

func (s NodeLimit) Apply(c *GraphConfig) {
	if s <= 0 {
		c.MaxNodes = 0
		return
	}
	c.MaxNodes = int(s)
}

func (s DepthLimit) Apply(c *GraphConfig) {
	if s <= 0 {
		c.MaxDepth = 0
		return
	}
	c.MaxDepth = int(s)
}
