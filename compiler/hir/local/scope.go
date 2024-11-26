package local

import "github.com/kkkunny/Sim/compiler/hir"

// Scope 作用域
type Scope interface {
	Package() *hir.Package
	SetIdent(name string, ident any) bool
	GetIdent(name string, allowLinkedPkgs ...bool) (any, bool)
}
