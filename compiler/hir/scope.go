package hir

import "github.com/kkkunny/stl/container/optional"

type Scope interface {
	Parent() (Scope, bool)
	AddValue(v Ident)
	Lookup(name string) (Ident, bool)
}

type LocalScope interface {
	Scope
	SetFuncType(f *FuncType)
	FuncType() *FuncType
}

type PkgScope struct {
	values map[string]Ident
}

func NewPkgScope() *PkgScope {
	return &PkgScope{values: make(map[string]Ident)}
}

func (s *PkgScope) Parent() (Scope, bool) {
	return nil, false
}

func (s *PkgScope) AddValue(v Ident) {
	s.values[v.GetName()] = v
}

func (s *PkgScope) Lookup(name string) (Ident, bool) {
	v, ok := s.values[name]
	return v, ok
}

type BlockScope struct {
	parent   Scope
	funcType optional.Optional[*FuncType]
	values   map[string]Ident
}

func NewBlockScope(p Scope) *BlockScope {
	return &BlockScope{parent: p, values: make(map[string]Ident)}
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

func (s *BlockScope) Lookup(name string) (Ident, bool) {
	v, ok := s.values[name]
	if ok {
		return v, true
	}
	return s.parent.Lookup(name)
}
