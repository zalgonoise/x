package attr

type Attr interface {
	Key() string
	Value() any
	WithKey(key string) Attr
	WithValue(value any) Attr
}

func NewAttr[T any](key string, value T) Attr {
	if err, ok := (any)(value).(error); ok {
		return ErrAttr(key, err)
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
	return NewAttr(key, a.value)
}

func (a attr[T]) WithValue(value any) Attr {
	if v, ok := (value).(T); ok {
		return NewAttr(a.key, v)
	}
	return nil
}
