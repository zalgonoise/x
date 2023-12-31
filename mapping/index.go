package mapping

type Index[K comparable, T any] struct {
	Keys []K

	zero   T
	values map[K]T
}

func (i Index[K, T]) Get(key K) (T, bool) {
	var zero K

	if key == zero || len(i.values) == 0 {
		return i.zero, false
	}

	value, ok := i.values[key]
	if !ok {
		return i.zero, false
	}

	return value, true
}
