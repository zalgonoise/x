package mapping

type Table[K comparable, T any] struct {
	zero   T
	values map[K]T
}

func (t Table[K, T]) Get(key K) (T, bool) {
	var zero K

	if key == zero || len(t.values) == 0 {
		return t.zero, false
	}

	value, ok := t.values[key]
	if !ok {
		return t.zero, false
	}

	return value, true
}
