package options

type intID int

var (
	ConfAdjList Setting = MultiOption(
		GraphList,  // init type
		Unweighted, // adjancy lists do not hold weights
	)
)
