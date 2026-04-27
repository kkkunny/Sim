package scopes

import (
	"github.com/kkkunny/stl/container/optional"
	"github.com/kkkunny/stl/container/set"

	"github.com/kkkunny/Sim/compiler/hir/stmts"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

type Scope interface {
	Package() *PkgScope
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

type PkgScope struct {
	Name      string
	externals map[string]*PkgScope
	values    map[string]stmts.Ident
	usedValue set.Set[stmts.Ident]
	types     map[string]*stmts.TypeDef
}

func NewPkgScope(name string) *PkgScope {
	return &PkgScope{
		Name:      name,
		externals: make(map[string]*PkgScope),
		values:    make(map[string]stmts.Ident),
		usedValue: set.StdHashSetWith[stmts.Ident](),
		types:     make(map[string]*stmts.TypeDef),
	}
}

func (s *PkgScope) Package() *PkgScope {
	return s
}

func (s *PkgScope) AddExternal(name string, pkg *PkgScope) {
	s.externals[name] = pkg
}

func (s *PkgScope) LookupPkg(name string) (*PkgScope, bool) {
	p, ok := s.externals[name]
	return p, ok
}

func (s *PkgScope) Parent() (Scope, bool) {
	return nil, false
}

func (s *PkgScope) AddValue(v stmts.Ident) {
	s.values[v.GetName()] = v
}

func (s *PkgScope) LookupValue(name string) (v stmts.Ident, ok bool) {
	defer func() {
		if ok {
			s.usedValue.Add(v)
		}
	}()
	return s.LocalLookupValue(name)
}

func (s *PkgScope) LocalLookupValue(name string) (v stmts.Ident, ok bool) {
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

func (s *PkgScope) AddType(name string, td *stmts.TypeDef) {
	s.types[name] = td
}

func (s *PkgScope) LookupType(name string) (*stmts.TypeDef, bool) {
	t, ok := s.types[name]
	return t, ok
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
