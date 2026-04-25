package scopes

import (
	"github.com/kkkunny/stl/container/optional"
	"github.com/kkkunny/stl/container/set"

	"github.com/kkkunny/Sim/compiler/hir/stmts"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

type Scope interface {
	Package() *PkgScope
	Parent() (Scope, bool)
	AddValue(v stmts.Ident)
	Lookup(name string) (stmts.Ident, bool)
	LocalLookup(name string) (stmts.Ident, bool)
	Values() map[string]stmts.Ident
	UsedValues() []stmts.Ident
}

type LocalScope interface {
	Scope
	SetFuncType(f types.FuncType)
	FuncType() types.FuncType
}

type PkgScope struct {
	values    map[string]stmts.Ident
	usedValue set.Set[stmts.Ident]
}

func NewPkgScope() *PkgScope {
	return &PkgScope{
		values:    make(map[string]stmts.Ident),
		usedValue: set.StdHashSetWith[stmts.Ident](),
	}
}

func (s *PkgScope) Package() *PkgScope {
	return s
}

func (s *PkgScope) Parent() (Scope, bool) {
	return nil, false
}

func (s *PkgScope) AddValue(v stmts.Ident) {
	s.values[v.GetName()] = v
}

func (s *PkgScope) Lookup(name string) (v stmts.Ident, ok bool) {
	defer func() {
		if ok {
			s.usedValue.Add(v)
		}
	}()
	return s.LocalLookup(name)
}

func (s *PkgScope) LocalLookup(name string) (v stmts.Ident, ok bool) {
	defer func() {
		if ok {
			s.usedValue.Add(v)
		}
	}()
	v, ok = s.values[name]
	return v, ok
}

func (s *PkgScope) Values() map[string]stmts.Ident {
	return s.values
}

func (s *PkgScope) UsedValues() []stmts.Ident {
	return s.usedValue.ToSlice()
}

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

func (s *BlockScope) Package() *PkgScope {
	return s.parent.Package()
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

func (s *BlockScope) Lookup(name string) (v stmts.Ident, ok bool) {
	defer func() {
		if ok {
			s.usedValue.Add(v)
		}
	}()
	if v, ok = s.LocalLookup(name); ok {
		return v, true
	}
	return s.parent.Lookup(name)
}

func (s *BlockScope) LocalLookup(name string) (v stmts.Ident, ok bool) {
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
