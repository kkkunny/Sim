package scopes

import (
	stlmaps "github.com/kkkunny/stl/container/maps"
	"github.com/kkkunny/stl/container/set"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/globals"
	"github.com/kkkunny/Sim/compiler/hir/locals"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

type PkgScope struct {
	Name string

	externals map[string]*PkgScope
	includes  set.Set[*PkgScope]

	values    map[string]hir.Ident
	usedValue set.Set[hir.Ident]
	types     map[string]types.CustomType
	binds     map[*globals.TypeDef]map[string]*locals.Let
}

func NewPkgScope(name string) *PkgScope {
	return &PkgScope{
		Name: name,

		externals: make(map[string]*PkgScope),
		includes:  set.StdHashSetWith[*PkgScope](),

		values:    make(map[string]hir.Ident),
		usedValue: set.StdHashSetWith[hir.Ident](),
		types:     make(map[string]types.CustomType),
		binds:     make(map[*globals.TypeDef]map[string]*locals.Let),
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

func (s *PkgScope) AddValue(v hir.Ident) {
	s.values[v.GetName()] = v
}

func (s *PkgScope) LookupValue(name string) (v hir.Ident, ok bool) {
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

func (s *PkgScope) Values() map[string]hir.Ident {
	return s.values
}

func (s *PkgScope) UsedValues() []hir.Ident {
	return s.usedValue.ToSlice()
}

func (s *PkgScope) AddType(name string, ct types.CustomType) {
	s.types[name] = ct
}

func (s *PkgScope) LookupType(name string) (types.CustomType, bool) {
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

func (s *PkgScope) AddBind(typeDef *globals.TypeDef, let *locals.Let) {
	if !stlmaps.ContainKey(s.binds, typeDef) {
		s.binds[typeDef] = make(map[string]*locals.Let)
	}
	s.binds[typeDef][let.Name] = let
}

func (s *PkgScope) LookupBind(typeDef *globals.TypeDef, name string) (*locals.Let, bool) {
	typeBinds, ok := s.binds[typeDef]
	if !ok {
		return nil, false
	}
	let, ok := typeBinds[name]
	return let, ok
}
