package hir

type Ident interface {
	Expr
	GetName() string
}
