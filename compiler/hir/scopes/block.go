package scopes

import (
	"github.com/kkkunny/stl/container/optional"
	"github.com/kkkunny/stl/container/set"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/globals"
	"github.com/kkkunny/Sim/compiler/hir/locals"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

type BlockScope struct {
	parent    Scope
	funcType  optional.Optional[types.FuncType]
	values    map[string]hir.Ident
	usedValue set.Set[hir.Ident]
}

func NewBlockScope(p Scope) *BlockScope {
	return &BlockScope{
		parent:    p,
		values:    make(map[string]hir.Ident),
		usedValue: set.StdHashSetWith[hir.Ident](),
	}
}

func (s *BlockScope) Root() *PkgScope {
	return s.parent.Root()
}

func (s *BlockScope) LookupPkg(name string) (*PkgScope, bool) {
	return s.parent.LookupPkg(name)
}

func (s *BlockScope) Parent() (Scope, bool) {
	return s.parent, true
}

func (s *BlockScope) SetFuncType(f types.FuncType) {
	s.funcType = optional.Some(f)
}

func (s *BlockScope) FuncType() types.FuncType {
	if f, ok := s.funcType.Value(); ok {
		return f
	}
	if ls, ok := s.parent.(LocalScope); ok {
		return ls.FuncType()
	}
	panic("unreachable")
}

func (s *BlockScope) AddValue(v hir.Ident) {
	s.values[v.GetName()] = v
}

func (s *BlockScope) LookupValue(name string) (v hir.Ident, ok bool) {
	defer func() {
		if ok {
			s.usedValue.Add(v)
		}
	}()
	if v, ok = s.values[name]; ok {
		return v, true
	}
	return s.parent.LookupValue(name)
}

func (s *BlockScope) Values() map[string]hir.Ident {
	return s.values
}

func (s *BlockScope) UsedValues() []hir.Ident {
	return s.usedValue.ToSlice()
}

func (s *BlockScope) LookupType(name string) (types.CustomType, bool) {
	return s.parent.LookupType(name)
}

func (s *BlockScope) AddBind(typeDef *globals.TypeDef, let *locals.Let) {
	s.parent.AddBind(typeDef, let)
}

func (s *BlockScope) LookupBind(typeDef *globals.TypeDef, name string) (*locals.Let, bool) {
	return s.parent.LookupBind(typeDef, name)
}
