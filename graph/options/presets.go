package options

type intID int

var (
	CfgAdjacencyList Setting = MultiOption(
		GraphList,  // init type
		Unweighted, // adjancy lists do not hold weights
	)
)
