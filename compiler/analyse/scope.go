package analyse

import (
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/either"
	"github.com/kkkunny/stl/container/hashmap"
	"github.com/kkkunny/stl/container/linkedhashset"

	"github.com/kkkunny/Sim/hir"
)

// 作用域
type _Scope interface {
	SetValue(name string, v hir.Ident) bool
	GetValue(pkg, name string) (hir.Ident, bool)
}

// 包作用域
type _PkgScope struct {
	pkg     hir.Package
	externs hashmap.HashMap[string, *_PkgScope]
	links   linkedhashset.LinkedHashSet[*_PkgScope]

	valueDefs hashmap.HashMap[string, hir.Ident]
	typeDefs  hashmap.HashMap[string, hir.GlobalType]
	traitDefs hashmap.HashMap[string, *hir.Trait]
}

func _NewPkgScope(pkg hir.Package) *_PkgScope {
	return &_PkgScope{
		pkg: pkg,
	}
}

func (self *_PkgScope) SetValue(name string, v hir.Ident) bool {
	if _, ok := self.getValue(name); ok {
		return false
	}
	self.valueDefs.Set(name, v)
	return true
}

func (self *_PkgScope) getLocalValue(name string) (hir.Ident, bool) {
	return self.valueDefs.Get(name), self.valueDefs.ContainKey(name)
}

func (self *_PkgScope) getValue(name string) (hir.Ident, bool) {
	v, ok := self.getLocalValue(name)
	if ok {
		return v, true
	}
	for iter := self.links.Iterator(); iter.Next(); {
		v, ok := iter.Value().getLocalValue(name)
		if ok && v.(hir.Global).GetPublic() {
			return v, true
		}
	}
	return nil, false
}

func (self *_PkgScope) GetValue(pkg, name string) (hir.Ident, bool) {
	if pkg == "" {
		return self.getValue(name)
	}
	pkgScope := self.externs.Get(pkg)
	if pkgScope == nil {
		return nil, false
	}
	v, ok := pkgScope.getLocalValue(name)
	if !ok || !v.(hir.Global).GetPublic() {
		return nil, false
	}
	return v, true
}

func (self *_PkgScope) SetTypeDef(td hir.GlobalType) bool {
	if _, ok := self.getTypeDef(td.GetName()); ok {
		return false
	}
	if _, ok := self.getTrait(td.GetName()); ok {
		return false
	}
	self.typeDefs.Set(td.GetName(), td)
	return true
}

func (self *_PkgScope) getLocalTypeDef(name string) (hir.GlobalType, bool) {
	return self.typeDefs.Get(name), self.typeDefs.ContainKey(name)
}

func (self *_PkgScope) getTypeDef(name string) (hir.GlobalType, bool) {
	td, ok := self.getLocalTypeDef(name)
	if ok {
		return td, true
	}
	for iter := self.links.Iterator(); iter.Next(); {
		st, ok := iter.Value().getLocalTypeDef(name)
		if ok && st.GetPublic() {
			return st, true
		}
	}
	return nil, false
}

func (self *_PkgScope) GetTypeDef(pkg, name string) (hir.GlobalType, bool) {
	if pkg == "" {
		return self.getTypeDef(name)
	}
	pkgScope := self.externs.Get(pkg)
	if pkgScope == nil {
		return nil, false
	}
	td, ok := pkgScope.getLocalTypeDef(name)
	if !ok || !td.GetPublic() {
		return nil, false
	}
	return td, true
}

func (self *_PkgScope) SetTrait(trait *hir.Trait) bool {
	if _, ok := self.getValue(trait.Name); ok {
		return false
	}
	if _, ok := self.getTypeDef(trait.Name); ok {
		return false
	}
	self.traitDefs.Set(trait.Name, trait)
	return true
}

func (self *_PkgScope) getLocalTrait(name string) (*hir.Trait, bool) {
	return self.traitDefs.Get(name), self.traitDefs.ContainKey(name)
}

func (self *_PkgScope) getTrait(name string) (*hir.Trait, bool) {
	v, ok := self.getLocalTrait(name)
	if ok {
		return v, true
	}
	for iter := self.links.Iterator(); iter.Next(); {
		v, ok := iter.Value().getLocalTrait(name)
		if ok && v.GetPublic() {
			return v, true
		}
	}
	return nil, false
}

func (self *_PkgScope) GetTrait(pkg, name string) (*hir.Trait, bool) {
	if pkg == "" {
		return self.getTrait(name)
	}
	pkgScope := self.externs.Get(pkg)
	if pkgScope == nil {
		return nil, false
	}
	v, ok := pkgScope.getLocalTrait(name)
	if !ok || !v.GetPublic() {
		return nil, false
	}
	return v, true
}

func (self *_PkgScope) Isize() hir.GlobalType { return stlbasic.IgnoreWith(self.getTypeDef("isize")) }
func (self *_PkgScope) I8() hir.GlobalType    { return stlbasic.IgnoreWith(self.getTypeDef("i8")) }
func (self *_PkgScope) I16() hir.GlobalType   { return stlbasic.IgnoreWith(self.getTypeDef("i16")) }
func (self *_PkgScope) I32() hir.GlobalType   { return stlbasic.IgnoreWith(self.getTypeDef("i32")) }
func (self *_PkgScope) I64() hir.GlobalType   { return stlbasic.IgnoreWith(self.getTypeDef("i64")) }
func (self *_PkgScope) Usize() hir.GlobalType { return stlbasic.IgnoreWith(self.getTypeDef("usize")) }
func (self *_PkgScope) U8() hir.GlobalType    { return stlbasic.IgnoreWith(self.getTypeDef("u8")) }
func (self *_PkgScope) U16() hir.GlobalType   { return stlbasic.IgnoreWith(self.getTypeDef("u16")) }
func (self *_PkgScope) U32() hir.GlobalType   { return stlbasic.IgnoreWith(self.getTypeDef("u32")) }
func (self *_PkgScope) U64() hir.GlobalType   { return stlbasic.IgnoreWith(self.getTypeDef("u64")) }
func (self *_PkgScope) F32() hir.GlobalType   { return stlbasic.IgnoreWith(self.getTypeDef("f32")) }
func (self *_PkgScope) F64() hir.GlobalType   { return stlbasic.IgnoreWith(self.getTypeDef("f64")) }
func (self *_PkgScope) Bool() hir.GlobalType  { return stlbasic.IgnoreWith(self.getTypeDef("bool")) }
func (self *_PkgScope) Str() hir.GlobalType   { return stlbasic.IgnoreWith(self.getTypeDef("str")) }
func (self *_PkgScope) Default() *hir.Trait   { return stlbasic.IgnoreWith(self.getTrait("Default")) }
func (self *_PkgScope) Add() *hir.Trait       { return stlbasic.IgnoreWith(self.getTrait("Add")) }
func (self *_PkgScope) Sub() *hir.Trait       { return stlbasic.IgnoreWith(self.getTrait("Sub")) }
func (self *_PkgScope) Mul() *hir.Trait       { return stlbasic.IgnoreWith(self.getTrait("Mul")) }
func (self *_PkgScope) Div() *hir.Trait       { return stlbasic.IgnoreWith(self.getTrait("Div")) }
func (self *_PkgScope) Rem() *hir.Trait       { return stlbasic.IgnoreWith(self.getTrait("Rem")) }
func (self *_PkgScope) And() *hir.Trait       { return stlbasic.IgnoreWith(self.getTrait("And")) }
func (self *_PkgScope) Or() *hir.Trait        { return stlbasic.IgnoreWith(self.getTrait("Or")) }
func (self *_PkgScope) Xor() *hir.Trait       { return stlbasic.IgnoreWith(self.getTrait("Xor")) }
func (self *_PkgScope) Shl() *hir.Trait       { return stlbasic.IgnoreWith(self.getTrait("Shl")) }
func (self *_PkgScope) Shr() *hir.Trait       { return stlbasic.IgnoreWith(self.getTrait("Shr")) }
func (self *_PkgScope) Eq() *hir.Trait        { return stlbasic.IgnoreWith(self.getTrait("Eq")) }
func (self *_PkgScope) Lt() *hir.Trait        { return stlbasic.IgnoreWith(self.getTrait("Lt")) }
func (self *_PkgScope) Gt() *hir.Trait        { return stlbasic.IgnoreWith(self.getTrait("Gt")) }
func (self *_PkgScope) Land() *hir.Trait      { return stlbasic.IgnoreWith(self.getTrait("Land")) }
func (self *_PkgScope) Lor() *hir.Trait       { return stlbasic.IgnoreWith(self.getTrait("Lor")) }
func (self *_PkgScope) Neg() *hir.Trait       { return stlbasic.IgnoreWith(self.getTrait("Neg")) }
func (self *_PkgScope) Not() *hir.Trait       { return stlbasic.IgnoreWith(self.getTrait("Not")) }

// 本地作用域
type _LocalScope interface {
	_Scope
	GetParent() _Scope
	GetFuncScope() *_FuncScope
	GetPkgScope() *_PkgScope
	GetFunc() hir.CallableDef
	SetLoop(loop hir.Loop)
	GetLoop() hir.Loop
}

// 函数作用域
type _FuncScope struct {
	_BlockScope
	parent either.Either[*_PkgScope, _LocalScope]
	def    hir.CallableDef
}

func _NewFuncScope(p *_PkgScope, def hir.GlobalFuncOrMethod) *_FuncScope {
	self := &_FuncScope{
		parent: either.Left[*_PkgScope, _LocalScope](p),
		def:    def,
	}
	self._BlockScope = *_NewBlockScope(self)
	return self
}

func _NewLambdaScope(p _LocalScope, def hir.CallableDef) *_FuncScope {
	self := &_FuncScope{
		parent: either.Right[*_PkgScope, _LocalScope](p),
		def:    def,
	}
	self._BlockScope = *_NewBlockScope(self)
	return self
}

func (self *_FuncScope) SetValue(name string, v hir.Ident) bool {
	return self._BlockScope.SetValue(name, v)
}

func (self *_FuncScope) GetValue(pkg, name string) (hir.Ident, bool) {
	if pkg == "" && self.values.ContainKey(name) {
		return self.values.Get(name), true
	}
	return self.GetParent().GetValue(pkg, name)
}

func (self *_FuncScope) GetParent() _Scope {
	if pkg, ok := self.parent.Left(); ok {
		return pkg
	}
	return stlbasic.IgnoreWith(self.parent.Right())
}

func (self *_FuncScope) GetFuncScope() *_FuncScope {
	return self
}

func (self *_FuncScope) GetPkgScope() *_PkgScope {
	if pkg, ok := self.parent.Left(); ok {
		return pkg
	}
	return stlbasic.IgnoreWith(self.parent.Right()).GetPkgScope()
}

func (self *_FuncScope) GetFunc() hir.CallableDef {
	return self.def
}

// 代码块作用域
type _BlockScope struct {
	parent _LocalScope
	values hashmap.HashMap[string, hir.Ident]
	loop   hir.Loop
}

func _NewBlockScope(p _LocalScope) *_BlockScope {
	return &_BlockScope{
		parent: p,
		values: hashmap.NewHashMap[string, hir.Ident](),
	}
}

func (self *_BlockScope) SetValue(name string, v hir.Ident) bool {
	self.values.Set(name, v)
	return true
}

func (self *_BlockScope) GetValue(pkg, name string) (hir.Ident, bool) {
	if pkg != "" {
		return self.parent.GetValue(pkg, name)
	}
	if self.values.ContainKey(name) {
		return self.values.Get(name), true
	}
	return self.parent.GetValue("", name)
}

func (self *_BlockScope) GetParent() _Scope {
	return self.parent
}

func (self *_BlockScope) GetFuncScope() *_FuncScope {
	return self.parent.GetFuncScope()
}

func (self *_BlockScope) GetPkgScope() *_PkgScope {
	return self.parent.GetPkgScope()
}

func (self *_BlockScope) GetFunc() hir.CallableDef {
	return self.parent.GetFunc()
}

func (self *_BlockScope) SetLoop(loop hir.Loop) {
	self.loop = loop
}

func (self *_BlockScope) GetLoop() hir.Loop {
	if self.loop != nil {
		return self.loop
	}
	if self.parent != nil {
		return self.parent.GetLoop()
	}
	return nil
}
