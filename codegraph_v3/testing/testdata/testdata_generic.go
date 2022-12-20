package testing

func First[T any](slice []T) T {
	return slice[0]
}

func FirstAndLast[T any](slice ...T) (T, T) {
	return slice[0], slice[len(slice)-1]
}

type Slicer[T any] interface {
	First() T
	FirstAndLast() (T, T)
}

type slicer[T any] struct {
	slice []T
	len   int
	idx   int
}

func (s *slicer[T]) First() T {
	return s.slice[0]
}

func (s *slicer[T]) FirstAndLast() (T, T) {
	return s.slice[0], s.slice[s.len-1]
}
