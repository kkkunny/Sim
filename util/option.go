package util

type Option[T any] struct {
	v *T
}

func Some[T any](v T) Option[T] {
	return Option[T]{v: &v}
}

func None[T any]() Option[T] {
	return Option[T]{v: nil}
}

func Optional[T any](v *T) Option[T] {
	if v == nil {
		return None[T]()
	} else {
		return Some(*v)
	}
}

func (self Option[T]) IsSome() bool {
	return self.v != nil
}

func (self Option[T]) IsNone() bool {
	return self.v == nil
}

func (self Option[T]) Value() (T, bool) {
	if self.IsNone() {
		var v T
		return v, false
	}
	return *self.v, true
}
