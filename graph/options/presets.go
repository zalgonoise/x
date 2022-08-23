package options

var (
	NoType = struct{}{}
)

type intID int

var (
	CfgAdjacencyList Setting = MultiOption(
		GraphList,  // init type
		Unweighted, // adjancy lists do not hold weights
	)
	CfgLinkedList Setting = MultiOption(
		GraphList,      // init type
		Unweighted,     // adjancy lists do not hold weights
		NonCyclical,    // linked lists have no edges
		NonDirectional, // linked lists have no edges
		MaxNodes(1),    // linked lists only have one node
	)
)
