package mode

type Mode int

const (
	Monitor Mode = iota
	Record
	Filter
)

var modeKeys = map[Mode]string{
	Monitor: "monitor",
	Record:  "record",
	Filter:  "filter",
}
var modeValues = map[string]Mode{
	"monitor": Monitor,
	"record":  Record,
	"filter":  Filter,
}
