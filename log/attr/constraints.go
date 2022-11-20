package attr

import "time"

// IntRestriction is a constraint that only accepts int values
type IntRestriction interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// UintRestriction is a constraint that only accepts uint values
type UintRestriction interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// FloatRestriction is a constraint that only accepts float values
type FloatRestriction interface {
	~float32 | ~float64
}

// ComplexRestriction is a constraint that only accepts complex values
type ComplexRestriction interface {
	~complex64 | ~complex128
}

// NumberRestriction is a constraint that only accepts number values,
// as a combination of other constraints
type NumberRestriction interface {
	IntRestriction | UintRestriction | FloatRestriction | ComplexRestriction
}

// CharRestriction is a constraint that only accepts stringifiable tokens
type CharRestriction interface {
	~string | ~byte | ~rune | ~[]byte | ~[]rune
}

// BoolRestriction is a constraint that only accepts booleans
type BoolRestriction interface {
	~bool
}

// TimeRestriction is a constraint that only accepts time.Time values
type TimeRestriction interface {
	time.Time
}

// TextRestriction is a constraint that only accepts text values,
// as a combination of other constraints
type TextRestriction interface {
	CharRestriction | BoolRestriction | TimeRestriction
}
