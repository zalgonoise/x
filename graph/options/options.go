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
)

const (
	GraphList TypeSetting = iota
	GraphMatrix
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

func (s WeightAsDistance) Apply(c *GraphConfig) {
	if s == DistanceWeight {
		c.WeightAsDistance = true
		return
	}
}
