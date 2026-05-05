package scopes

import (
	"github.com/kkkunny/stl/container/optional"
	"github.com/kkkunny/stl/container/set"

	"github.com/kkkunny/Sim/compiler/hir/stmts"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

type BlockScope struct {
	parent    Scope
	funcType  optional.Optional[types.FuncType]
	values    map[string]stmts.Ident
	usedValue set.Set[stmts.Ident]
}

func NewBlockScope(p Scope) *BlockScope {
	return &BlockScope{
		parent:    p,
		values:    make(map[string]stmts.Ident),
		usedValue: set.StdHashSetWith[stmts.Ident](),
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

func (s *BlockScope) AddValue(v stmts.Ident) {
	s.values[v.GetName()] = v
}

func (s *BlockScope) LookupValue(name string) (v stmts.Ident, ok bool) {
	defer func() {
		if ok {
			s.usedValue.Add(v)
		}
	}()
	if v, ok = s.LocalLookupValue(name); ok {
		return v, true
	}
	return s.parent.LookupValue(name)
}

func (s *BlockScope) LocalLookupValue(name string) (v stmts.Ident, ok bool) {
	defer func() {
		if ok {
			s.usedValue.Add(v)
		}
	}()
	v, ok = s.values[name]
	return v, ok
}

func (s *BlockScope) Values() map[string]stmts.Ident {
	return s.values
}

func (s *BlockScope) UsedValues() []stmts.Ident {
	return s.usedValue.ToSlice()
}

func (s *BlockScope) LookupType(name string) (*stmts.TypeDef, bool) {
	return s.parent.LookupType(name)
}
