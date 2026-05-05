package scopes

import (
	"github.com/kkkunny/Sim/compiler/hir/stmts"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

type Scope interface {
	Root() *PkgScope
	LookupPkg(name string) (*PkgScope, bool)
	Parent() (Scope, bool)
	AddValue(v stmts.Ident)
	LookupValue(name string) (stmts.Ident, bool)
	LocalLookupValue(name string) (stmts.Ident, bool)
	Values() map[string]stmts.Ident
	UsedValues() []stmts.Ident
	LookupType(name string) (*stmts.TypeDef, bool)
}

type LocalScope interface {
	Scope
	SetFuncType(f types.FuncType)
	FuncType() types.FuncType
}
