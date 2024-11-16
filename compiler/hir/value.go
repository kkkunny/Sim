package hir

type Value interface {
	Type() Type
	Mutable() bool
	Storable() bool
}
