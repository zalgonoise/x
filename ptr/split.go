package ptr

import "unsafe"

const (
	by2 int = iota + 2
	by3
	by4
	by5
)

// Split2 breaks the input slice into one with sets of two items each:
//
// input: []int{0, 1, 2, 3, 4}
// output: [][2]int{[2]int{0, 1}, [2]int{2, 3}, [2]int{4, 0}}
func Split2[T any](data []T) [][2]T {
	const size = by2

	ln := len(data) / size
	cutoff := len(data) % size

	newData := *(*[][size]T)(unsafe.Pointer(&data))

	if cutoff == 0 && ln > 0 {
		return newData[:ln]
	}

	ln++
	newData = newData[:ln]

	for i := size - 1; i >= cutoff; i-- {
		newData[len(newData)-1][i] = *(new(T))
	}

	return newData
}

// Split3 breaks the input slice into one with sets of three items each:
//
// input: []int{0, 1, 2, 3, 4}
// output: [][3]int{[3]int{0, 1, 2}, [3]int{3, 4, 0}}
func Split3[T any](data []T) [][3]T {
	const size = by3

	ln := len(data) / size
	cutoff := len(data) % size

	newData := *(*[][size]T)(unsafe.Pointer(&data))

	if cutoff == 0 && ln > 0 {
		return newData[:ln]
	}

	ln++
	newData = newData[:ln]

	for i := size - 1; i >= cutoff; i-- {
		newData[len(newData)-1][i] = *(new(T))
	}

	return newData
}

// Split4 breaks the input slice into one with sets of four items each:
//
// input: []int{0, 1, 2, 3, 4}
// output: [][4]int{[4]int{0, 1, 2, 3}, [4]int{4, 0, 0, 0}}
func Split4[T any](data []T) [][4]T {
	const size = by4

	ln := len(data) / size
	cutoff := len(data) % size

	newData := *(*[][size]T)(unsafe.Pointer(&data))

	if cutoff == 0 && ln > 0 {
		return newData[:ln]
	}

	ln++
	newData = newData[:ln]

	for i := size - 1; i >= cutoff; i-- {
		newData[len(newData)-1][i] = *(new(T))
	}

	return newData
}

// Split5 breaks the input slice into one with sets of five items each:
//
// input: []int{0, 1, 2, 3}
// output: [][5]int{[5]int{0, 1, 2, 3, 0}}
func Split5[T any](data []T) [][5]T {
	const size = by5

	ln := len(data) / size
	cutoff := len(data) % size

	newData := *(*[][size]T)(unsafe.Pointer(&data))

	if cutoff == 0 && ln > 0 {
		return newData[:ln]
	}

	ln++
	newData = newData[:ln]

	for i := size - 1; i >= cutoff; i-- {
		newData[len(newData)-1][i] = *(new(T))
	}

	return newData
}
