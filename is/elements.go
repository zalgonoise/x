package is

import (
	"cmp"
	"slices"
)

// Contains checks if the input slice contains the input item, provided that they are of a comparable type.
func Contains[V comparable, S ~[]V](t T, item V, slice S) {
	if ok := contains(item, slice); ok {
		t.Logf(outputOKFormat, item)

		return
	}

	t.Logf(outputNOKFormat, item, slice)
	t.Fail()
}

// ElementsMatch asserts that the elements in the expected slice match the elements in the actual slice,
// regardless of their order. If order is important, either slices.Equal or EqualElements should be preferable options.
//
// Ex:
//   - expected: [1, 2, 3]
//   - actual: [3, 1, 2]
//   - result: OK
func ElementsMatch[V cmp.Ordered, S ~[]V](t T, expected, actual S) {
	if h, ok := t.(helper); ok {
		h.Helper()
	}

	if ok := elementsMatch(expected, actual); ok {
		t.Logf(outputOKFormat, expected)

		return
	}

	t.Logf(outputNOKFormat, expected, actual)
	t.Fail()
}

// EqualElements asserts that the elements match the ones in the actual slice, in both position and input.
// It is a shorthand for calling slices.Equal.
//
// Ex:
//   - expected: [1, 2, 3]
//   - actual: [1, 2, 3]
//   - result: OK
func EqualElements[V cmp.Ordered, S ~[]V](t T, expected, actual S) {
	if h, ok := t.(helper); ok {
		h.Helper()
	}

	if ok := slices.Equal(expected, actual); ok {
		t.Logf(outputOKFormat, expected)

		return
	}

	t.Logf(outputNOKFormat, expected, actual)
	t.Fail()
}

func contains[V comparable, S ~[]V](item V, slice S) bool {
	for i := range slice {
		if slice[i] == item {
			return true
		}
	}

	return false
}

func elementsMatch[V cmp.Ordered, S ~[]V](expected, actual S) bool {
	eCopy := make(S, len(expected))
	copy(eCopy, expected)

	aCopy := make(S, len(actual))
	copy(aCopy, actual)

	slices.Sort(eCopy)
	slices.Sort(aCopy)

	return slices.Equal(eCopy, aCopy)
}
