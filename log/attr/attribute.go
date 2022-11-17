package attr

type Attr interface {
	Key() string
	Value() any
	WithKey(key string) Attr
	WithValue(value any) Attr
}

func New[T any](key string, value T) Attr {
	if key == "" {
		return nil
	}
	return attr[T]{
		key:   key,
		value: value,
	}
}

type attr[T any] struct {
	key   string
	value T
}

func (a attr[T]) Key() string {
	return a.key
}

func (a attr[T]) Value() any {
	return a.value
}

func (a attr[T]) WithKey(key string) Attr {
	if key == "" {
		return nil
	}
	return New(key, a.value)
}

func (a attr[T]) WithValue(value any) Attr {
	if value == nil {
		return nil
	}

	v, ok := (value).(T)
	if !ok {
		return nil
	}
	return New(a.key, v)
}
