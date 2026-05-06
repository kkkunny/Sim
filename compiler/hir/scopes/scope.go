package scopes

import (
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/globals"
	"github.com/kkkunny/Sim/compiler/hir/locals"
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
	AddBind(typeDef *globals.TypeDef, let *locals.Let)
	LookupBind(typeDef *globals.TypeDef, name string) (*locals.Let, bool)
}

type LocalScope interface {
	Scope
	SetFuncType(f types.FuncType)
	FuncType() types.FuncType
}
