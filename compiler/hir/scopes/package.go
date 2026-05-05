package scopes

import (
	"github.com/kkkunny/stl/container/set"

	"github.com/kkkunny/Sim/compiler/hir/stmts"
)

type PkgScope struct {
	Name string

	externals map[string]*PkgScope
	includes  set.Set[*PkgScope]

	values    map[string]stmts.Ident
	usedValue set.Set[stmts.Ident]
	types     map[string]*stmts.TypeDef
}

func NewPkgScope(name string) *PkgScope {
	return &PkgScope{
		Name: name,

		externals: make(map[string]*PkgScope),
		includes:  set.StdHashSetWith[*PkgScope](),

		values:    make(map[string]stmts.Ident),
		usedValue: set.StdHashSetWith[stmts.Ident](),
		types:     make(map[string]*stmts.TypeDef),
	}
}

func (s *PkgScope) Root() *PkgScope {
	return s
}

func (s *PkgScope) AddExternal(name string, pkg *PkgScope) {
	s.externals[name] = pkg
}

func (s *PkgScope) AddInclude(pkg *PkgScope) {
	s.includes.Add(pkg)
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
	if v, ok = s.values[name]; ok {
		return v, true
	}
	for pkg := range s.includes.Iter() {
		if v, ok = pkg.LookupValue(name); ok {
			return v, true
		}
	}
	return v, false
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
	if t, ok := s.types[name]; ok {
		return t, true
	}
	for pkg := range s.includes.Iter() {
		if t, ok := pkg.LookupType(name); ok {
			return t, true
		}
	}
	return nil, false
}
