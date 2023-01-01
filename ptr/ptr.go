package ptr

import "reflect"

const (
	maxDepth = 1024
)

// From takes in a pointer to a type, and returns
// the value stored in the pointer, and a boolean on weather
// it is nil or not
//
// If the value is nil, it returns an initialized version of the
// type, regardless what value it is
func From[T any](ptr *T) (T, bool) {
	if ptr == nil {
		return *new(T), false
	}
	return *ptr, true
}

// Must will return the value stored in the input pointer to a type,
// if it is not nil
//
// If the value is nil, it returns an initialized version of the
// type, regardless what value it is
func Must[T any](ptr *T) T {
	if ptr == nil {
		return *new(T)
	}
	return *ptr
}

// To will return a pointer to the input value
func To[T any](value T) *T {
	return &value
}

// Copy will return a literal copy of the value stored in the input pointer
// as a pointer, too
func Copy[T any](ptr *T) *T {
	newT := new(T)
	*newT = *ptr
	return newT
}

// DeepCopy performs a copy on a type including its pointer data, provided that
// it contains only exported fields (including for nested structs and interfaces)
//
// NOTE: this is a slow and expensive operation; I would prefer if you didn't use it.
// It is only suitable when you want to copy data into a new object which contains
// nested pointers, and you really don't want to copy the items' elements one by one
//
// The input pointer `ptr` is copied into the output pointer `*T`, using reflection
// and as such it can only target exported fields.
//
// `depth` defines how deep will reflection go when finding nested objects (structs)
// to be copied -- otherwise it will be left as a zero value. The default maxDepth
// (of 1024) can be set using a negative number. Depths are also capped by this value.
func DeepCopy[T any](ptr *T, depth int) *T {
	n := new(T)
	t := reflect.ValueOf(n).Elem()
	f := reflect.ValueOf(ptr).Elem()

	if depth < 0 || depth > maxDepth {
		depth = maxDepth
	}
	if ok := setField(f, t, depth); !ok {
		return nil
	}
	return n
}

func setField(f, t reflect.Value, depth int) bool {
	switch f.Kind() {
	case reflect.Struct:
		for i := 0; i < f.NumField(); i++ {
			if depth > 0 {
				return setField(f.Field(i), t.Field(i), depth-1)
			}
		}
		return true
	case reflect.Pointer:
		if f.CanAddr() {
			t.Set(reflect.New(f.Type().Elem()))
			return setField(f.Elem(), t.Elem(), depth)
		}
	default:
		if f.IsValid() && f.CanSet() {
			t.Set(f)
			return true
		}
	}
	return false
}
