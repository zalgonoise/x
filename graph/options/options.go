package options

type Code uint8

const (
	Default Code = 0
	AdjList Code = 1 << iota
	NonDirectional
	NonCyclical
)

func Parse(opts Code) (isList, isNonDir, isNonCyc bool) {
	if opts == Default {
		return false, false, false
	}

	return opts&AdjList != 0,
		opts&NonDirectional != 0,
		opts&NonCyclical != 0

}
