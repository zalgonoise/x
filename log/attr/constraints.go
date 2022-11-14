package attr

import "time"

type IntRestriction interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}
type UintRestriction interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}
type FloatRestriction interface {
	~float32 | ~float64
}
type ComplexRestriction interface {
	~complex64 | ~complex128
}
type NumberRestriction interface {
	IntRestriction | UintRestriction | FloatRestriction | ComplexRestriction
}
type CharRestriction interface {
	~string | ~byte | ~rune | ~[]byte | ~[]rune
}
type BoolRestriction interface {
	~bool
}
type TimeRestriction interface {
	time.Time
}
type TextRestriction interface {
	CharRestriction | BoolRestriction | TimeRestriction
}
