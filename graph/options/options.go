package options

type Code uint8

const (
	Default Code = 0
	AdjList Code = 1 << iota
	NonDirectional
	NonCyclical
)

func IsList(code Code) bool {
	return code&AdjList != 0
}
func IsNonDirectional(code Code) bool {
	return code&NonDirectional != 0
}
func IsNonCyclical(code Code) bool {
	return code&NonCyclical != 0
}
