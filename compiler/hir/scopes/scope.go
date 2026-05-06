package scopes

import (
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

type Scope interface {
	Root() *PkgScope
	LookupPkg(name string) (*PkgScope, bool)
	Parent() (Scope, bool)
	AddValue(v hir.Ident)
	LookupValue(name string) (hir.Ident, bool)
	Values() map[string]hir.Ident
	UsedValues() []hir.Ident
	LookupType(name string) (types.CustomType, bool)
}

type LocalScope interface {
	Scope
	SetFuncType(f types.FuncType)
	FuncType() types.FuncType
}
