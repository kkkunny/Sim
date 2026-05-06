package scopes

import (
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/globals"
	"github.com/kkkunny/Sim/compiler/hir/locals"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

type TemporaryScope struct {
	parent Scope

	types map[string]types.CustomType
}

func NewTemporaryScope(p Scope) *TemporaryScope {
	return &TemporaryScope{
		parent: p,

		types: make(map[string]types.CustomType),
	}
}

func (s *TemporaryScope) Root() *PkgScope {
	return s.parent.Root()
}

func (s *TemporaryScope) LookupPkg(name string) (*PkgScope, bool) {
	return s.parent.LookupPkg(name)
}

func (s *TemporaryScope) Parent() (Scope, bool) {
	return s.parent, true
}

func (s *TemporaryScope) AddValue(v hir.Ident) {
	s.parent.AddValue(v)
}

func (s *TemporaryScope) LookupValue(name string) (v hir.Ident, ok bool) {
	return s.parent.LookupValue(name)
}

func (s *TemporaryScope) Values() map[string]hir.Ident {
	return s.parent.Values()
}

func (s *TemporaryScope) UsedValues() []hir.Ident {
	return s.parent.UsedValues()
}

func (s *TemporaryScope) AddType(name string, ct types.CustomType) {
	s.types[name] = ct
}

func (s *TemporaryScope) LookupType(name string) (types.CustomType, bool) {
	t, ok := s.types[name]
	if ok {
		return t, true
	}
	return s.parent.LookupType(name)
}

func (s *TemporaryScope) AddBind(typeDef *globals.TypeDef, let *locals.Let) {
	s.parent.AddBind(typeDef, let)
}

func (s *TemporaryScope) LookupBind(typeDef *globals.TypeDef, name string) (*locals.Let, bool) {
	return s.parent.LookupBind(typeDef, name)
}
