package local

// Scope 作用域
type Scope interface {
	SetIdent(name string, ident any) bool
	GetIdent(name string, allowLinkedPkgs ...bool) (any, bool)
}
