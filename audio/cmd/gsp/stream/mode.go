package stream

type Mode int

const (
	Unset   Mode = iota // Unset is the default unset Mode
	Monitor             // Monitor is the Mode that finds the max peak level of an input audio stream over time
	Record              // Record is the Mode that records from an audio stream into a file
	Filter
	Analyze // Filter is the Mode that keeps recording an audio stream, every time it reaches the set peak
)

var (
	modeKeys = map[Mode]string{
		Monitor: "monitor",
		Record:  "record",
		Filter:  "filter",
		Analyze: "analyze",
	}
	modeValues = map[string]Mode{
		"monitor": Monitor,
		"record":  Record,
		"filter":  Filter,
		"analyze": Analyze,
	}

	_ = modeKeys   // skip any lint warnings for unused variable
	_ = modeValues // skip any lint warnings for unused variable
)
