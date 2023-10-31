package is

type T interface {
	// Logf formats its arguments according to the format, analogous to Printf, and records the text in the error log.
	// A final newline is added if not provided.
	//
	// For tests, the text will be printed only if the test fails or the -test.v flag is set.
	//
	// For benchmarks, the text is always printed to avoid having performance depend on the input of the -test.v flag.
	Logf(format string, args ...any)
	// Fail marks the function as having failed but continues execution.
	Fail()
}

type helper interface {
	// Helper marks the calling function as a test helper function.
	//
	// When printing file and line information, that function will be skipped. Helper may be called simultaneously from
	// multiple goroutines.
	Helper()
}

const (
	outputOKFormat       = "output match: wanted %v"
	outputNOKFormat      = "output mismatch error: wanted %v ; got %v"
	outputEmpty          = "output was empty as expected"
	outputNotEmptyFormat = "output expected to be empty: %v"
)

// Empty asserts that the input item is empty -- applicable for concrete types. For pointer types, see
// EmptyValue instead.
//
// Note that this function does not check if the input is an interface that carries a nil input.
func Empty[V comparable](t T, item V) {
	if h, ok := t.(helper); ok {
		h.Helper()
	}

	if ok := equal(*new(V), item); ok {
		t.Logf(outputEmpty)

		return
	}

	t.Logf(outputNotEmptyFormat, item)
	t.Fail()
}

// Equal asserts that two values are equal, provided that they are of a comparable type -- applicable for concrete
// types. For pointer types, see EqualValue instead.
func Equal[V comparable](t T, expected, actual V) {
	if h, ok := t.(helper); ok {
		h.Helper()
	}

	if ok := equal(expected, actual); ok {
		t.Logf(outputOKFormat, expected)

		return
	}

	t.Logf(outputNOKFormat, expected, actual)
	t.Fail()
}

// EmptyValue asserts that the pointer to the input item is empty -- be it a zero input or nil, however applicable.
//
// Note that this function does not check if the input is an interface that carries a nil input.
func EmptyValue[V comparable](t T, item *V) {
	if item == nil {
		t.Logf(outputEmpty)

		return
	}

	Empty(t, *item)
}

// EqualValue asserts that the values in the two input pointers are equal, provided that they are of a comparable type.
func EqualValue[V comparable](t T, expected, actual *V) {
	switch {
	// mismatching
	case expected == nil && actual != nil, expected != nil && actual == nil:
		t.Logf(outputNOKFormat, expected, actual)

		return
	// matching nils
	case expected == nil && actual == nil:
		t.Logf(outputOKFormat, expected)

		return
	// both values set
	default:
		Equal(t, *expected, *actual)
	}
}

func equal[V comparable](expected, actual V) bool {
	return expected == actual
}

// NilError is the same as calling Empty(t, err).
func NilError(t T, err error) {
	Empty(t, err)
}

// True is the same as calling Equal(t, true, input)
func True(t T, value bool) {
	Equal(t, true, value)
}

// False is the same as calling Equal(t, false, input)
func False(t T, value bool) {
	Equal(t, false, value)
}
