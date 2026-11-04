package optional

type Optional[T any] struct {
	value   T
	defined bool
}

func Nullable[T any](value T) Optional[T] {
	return Optional[T]{value, true}
}

func Default[T any]() Optional[T] {
	return Optional[T]{}
}

func (optional Optional[T]) GetOr(fallback T) T {
	if optional.defined {
		return optional.value
	}
	return fallback
}

func (optional Optional[T]) isDefined() bool {
	return optional.defined
}
