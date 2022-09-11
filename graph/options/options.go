package options

type (
	TypeSetting      int
	DirectionSetting int
	CycleSetting     int
	WeightedEdges    int
	Mutability       int
	WritePrivilege   int
	NodeLimit        int
	DepthLimit       int
	WeightAsDistance int
	CrossGraphEdges  int
)

const (
	GraphList TypeSetting = iota
	GraphMatrix
	GraphKnowledge
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

const (
	LabelWeight WeightAsDistance = iota
	DistanceWeight
)

const (
	WithCrossGraphEdges CrossGraphEdges = iota
	NoCrossGraphEdges
)

func MaxNodes(v int) Setting {
	if v < 0 {
		v = 0
	}

	s := NodeLimit(v)
	return &s
}

func MaxDepth(v int) Setting {
	if v < 0 {
		v = 0
	}

	s := DepthLimit(v)
	return &s
}

func (s TypeSetting) Apply(c *GraphConfig) {
	c.GraphType = s
	if s == GraphMatrix {
		NoCrossGraphEdges.Apply(c)
		return
	}
	WithCrossGraphEdges.Apply(c)
}

func (s DirectionSetting) Apply(c *GraphConfig) {
	if s == NonDirectional {
		c.IsNonDirectional = true
		return
	}
	c.IsNonDirectional = false
}

func (s CycleSetting) Apply(c *GraphConfig) {
	if s == NonCyclical {
		c.IsNonCyclical = true
		return
	}
	c.IsNonCyclical = false
}

func (s WeightedEdges) Apply(c *GraphConfig) {
	if s == Unweighted {
		c.IsUnweighted = true
		return
	}
	c.IsUnweighted = false
}

func (s Mutability) Apply(c *GraphConfig) {
	if s == Immutable {
		c.Immutable = true
		return
	}
	c.Immutable = false
}

func (s WritePrivilege) Apply(c *GraphConfig) {
	if s == ReadOnly {
		c.ReadOnly = true
		return
	}
	c.ReadOnly = false
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

func (s WeightAsDistance) Apply(c *GraphConfig) {
	if s == DistanceWeight {
		c.WeightAsDistance = true
		return
	}
	c.WeightAsDistance = false
}

func (s CrossGraphEdges) Apply(c *GraphConfig) {
	if s == NoCrossGraphEdges {
		c.NoCrossGraphEdges = true
		return
	}
	c.NoCrossGraphEdges = false
}
