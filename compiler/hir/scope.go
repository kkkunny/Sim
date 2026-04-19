package hir

import (
	"github.com/kkkunny/stl/container/optional"
	"github.com/kkkunny/stl/container/set"
)

type Scope interface {
	Package() *PkgScope
	Parent() (Scope, bool)
	AddValue(v Ident)
	Lookup(name string) (Ident, bool)
	LocalLookup(name string) (Ident, bool)
	Values() map[string]Ident
	UsedValues() []Ident
}

type LocalScope interface {
	Scope
	SetFuncType(f *FuncType)
	FuncType() *FuncType
}

type PkgScope struct {
	values    map[string]Ident
	usedValue set.Set[Ident]
}

func NewPkgScope() *PkgScope {
	return &PkgScope{
		values:    make(map[string]Ident),
		usedValue: set.StdHashSetWith[Ident](),
	}
}

func (s *PkgScope) Package() *PkgScope {
	return s
}

func (s *PkgScope) Parent() (Scope, bool) {
	return nil, false
}

func (s *PkgScope) AddValue(v Ident) {
	s.values[v.GetName()] = v
}

func (s *PkgScope) Lookup(name string) (v Ident, ok bool) {
	defer func() {
		if ok {
			s.usedValue.Add(v)
		}
	}()
	return s.LocalLookup(name)
}

func (s *PkgScope) LocalLookup(name string) (v Ident, ok bool) {
	defer func() {
		if ok {
			s.usedValue.Add(v)
		}
	}()
	v, ok = s.values[name]
	return v, ok
}

func (s *PkgScope) Values() map[string]Ident {
	return s.values
}

func (s *PkgScope) UsedValues() []Ident {
	return s.usedValue.ToSlice()
}

type BlockScope struct {
	parent    Scope
	funcType  optional.Optional[*FuncType]
	values    map[string]Ident
	usedValue set.Set[Ident]
}

func NewBlockScope(p Scope) *BlockScope {
	return &BlockScope{
		parent:    p,
		values:    make(map[string]Ident),
		usedValue: set.StdHashSetWith[Ident](),
	}
}

func (s *BlockScope) Package() *PkgScope {
	return s.parent.Package()
}

func (s *BlockScope) Parent() (Scope, bool) {
	return s.parent, true
}

func (s *BlockScope) SetFuncType(f *FuncType) {
	s.funcType = optional.Some(f)
}

func (s *BlockScope) FuncType() *FuncType {
	if f, ok := s.funcType.Value(); ok {
		return f
	}
	if ls, ok := s.parent.(LocalScope); ok {
		return ls.FuncType()
	}
	panic("unreachable")
}

func (s *BlockScope) AddValue(v Ident) {
	s.values[v.GetName()] = v
}

func (s *BlockScope) Lookup(name string) (v Ident, ok bool) {
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

func (s *BlockScope) LocalLookup(name string) (v Ident, ok bool) {
	defer func() {
		if ok {
			s.usedValue.Add(v)
		}
	}()
	v, ok = s.values[name]
	return v, ok
}

func (s *BlockScope) Values() map[string]Ident {
	return s.values
}

func (s *BlockScope) UsedValues() []Ident {
	return s.usedValue.ToSlice()
}
