package hir

type Ident interface {
	GetName() string
	GetType() Type
}
