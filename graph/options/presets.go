package options

type intID int

var (
	CfgAdjencyList Setting = MultiOption(
		GraphList,  // init type
		Unweighted, // adjancy lists do not hold weights
	)
)
