package lex

type Item[T, V any] struct {
	Typ T
	Val []V
}

func NewItem[T, V any](itemType T, value ...V) Item[T, V] {
	return Item[T, V]{
		Typ: itemType,
		Val: value,
	}
}
