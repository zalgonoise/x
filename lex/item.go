package lex

// Item represents a set of any type of tokens identified by a comparable type
type Item[T comparable, V any] struct {
	Typ T
	Val []V
}

// NewItem creates an Item with type `T` and values `[]V`
func NewItem[T comparable, V any](itemType T, value ...V) Item[T, V] {
	return Item[T, V]{
		Typ: itemType,
		Val: value,
	}
}
