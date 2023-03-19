package osc

type Type uint8

const (
	SineWave Type = iota
	SquareWave
	TriangleWave
	SawtoothWave
)

type bitDepths interface {
	int8 | int16 | int32
}

type Oscillator[T bitDepths] func(buffer []T, freq, depth, sampleRate float64)
