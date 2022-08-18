package options

type intID int

var (
	ConfList Setting = MultiOption(
		GraphList,
		Unweighted,
		IDType(int(0)),
	)
)
