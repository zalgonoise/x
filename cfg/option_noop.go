package cfg

// NoOp is a placeholder option, suitable for any type, that can be used for, as an example, invalid input. The
// returned option will perform no changes to the input config and will return it as-is
type NoOp[T any] struct{}

func (NoOp[T]) apply(config T) T {
	return config
}
