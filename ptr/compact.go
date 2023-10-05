package ptr

import "unsafe"

// NonNil scans the input pointer types if they are nil, returning a slice containing only those that are non-nil
func NonNil[T any](values ...*T) []*T {
	if len(values) == 0 {
		return nil
	}

	output := make([]*T, 0, len(values))

	for i := range values {
		if values[i] != nil {
			output = append(output, values[i])
		}
	}

	return output
}

func Compact[T any](values ...*T) []*T {
	if len(values) == 0 {
		return nil
	}

	output := make([]*T, 0, len(values))
	cache := make(map[uintptr]struct{}, len(values))

	for i := range values {
		if values[i] != nil {
			ptr := uintptr(unsafe.Pointer(values[i]))
			if _, ok := cache[ptr]; ok {
				continue
			}

			cache[ptr] = struct{}{}
			output = append(output, values[i])
		}
	}

	return output
}
