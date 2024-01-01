package analyse

import (
	"strings"

	stlbasic "github.com/kkkunny/stl/basic"
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
	links     linkedhashset.LinkedHashSet[*_PkgScope]

	valueDefs hashmap.HashMap[string, hir.Ident]
	typeDefs  hashmap.HashMap[string, hir.TypeDef]
	genericFuncDefs hashmap.HashMap[string, *hir.GenericFuncDef]
	genericStructDefs hashmap.HashMap[string, *hir.GenericStructDef]
}

func _NewPkgScope(pkg hir.Package) *_PkgScope {
	return &_PkgScope{
		pkg:       pkg,
		externs:   hashmap.NewHashMap[string, *_PkgScope](),
		links:     linkedhashset.NewLinkedHashSet[*_PkgScope](),
		valueDefs: hashmap.NewHashMap[string, hir.Ident](),
		typeDefs:  hashmap.NewHashMap[string, hir.TypeDef](),
		genericFuncDefs: hashmap.NewHashMap[string, *hir.GenericFuncDef](),
		genericStructDefs: hashmap.NewHashMap[string, *hir.GenericStructDef](),
	}
}

func (self *_PkgScope) SetValue(name string, v hir.Ident) bool {
	if _, ok := self.getValue(name); ok {
		return false
	}
	if _, ok := self.getGenericFuncDef(name); ok {
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

func (self *_PkgScope) SetTypeDef(td hir.TypeDef) bool {
	if _, ok := self.getTypeDef(td.GetName()); ok {
		return false
	}
	if _, ok := self.getGenericStructDef(td.GetName()); ok {
		return false
	}
	self.typeDefs.Set(td.GetName(), td)
	return true
}

func (self *_PkgScope) getLocalTypeDef(name string) (hir.TypeDef, bool) {
	return self.typeDefs.Get(name), self.typeDefs.ContainKey(name)
}

func (self *_PkgScope) getTypeDef(name string) (hir.TypeDef, bool) {
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

func (self *_PkgScope) GetTypeDef(pkg, name string) (hir.TypeDef, bool) {
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

func (self *_PkgScope) SetGenericFuncDef(def *hir.GenericFuncDef) bool {
	if _, ok := self.getValue(def.Name); ok {
		return false
	}
	if _, ok := self.getGenericFuncDef(def.Name); ok {
		return false
	}
	self.genericFuncDefs.Set(def.Name, def)
	return true
}

func (self *_PkgScope) getLocalGenericFuncDef(name string) (*hir.GenericFuncDef, bool) {
	return self.genericFuncDefs.Get(name), self.genericFuncDefs.ContainKey(name)
}

func (self *_PkgScope) getGenericFuncDef(name string) (*hir.GenericFuncDef, bool) {
	def, ok := self.getLocalGenericFuncDef(name)
	if ok {
		return def, true
	}
	for iter := self.links.Iterator(); iter.Next(); {
		def, ok := iter.Value().getLocalGenericFuncDef(name)
		if ok && def.GetPublic() {
			return def, true
		}
	}
	return nil, false
}

func (self *_PkgScope) GetGenericFuncDef(pkg, name string) (*hir.GenericFuncDef, bool) {
	if pkg == "" {
		return self.getGenericFuncDef(name)
	}
	pkgScope := self.externs.Get(pkg)
	if pkgScope == nil {
		return nil, false
	}
	def, ok := pkgScope.getLocalGenericFuncDef(name)
	if !ok || !def.GetPublic() {
		return nil, false
	}
	return def, true
}

func (self *_PkgScope) SetGenericStructDef(def *hir.GenericStructDef) bool {
	if _, ok := self.getTypeDef(def.Name); ok {
		return false
	}
	if _, ok := self.getGenericStructDef(def.Name); ok {
		return false
	}
	self.genericStructDefs.Set(def.Name, def)
	return true
}

func (self *_PkgScope) getLocalGenericStructDef(name string) (*hir.GenericStructDef, bool) {
	return self.genericStructDefs.Get(name), self.genericStructDefs.ContainKey(name)
}

func (self *_PkgScope) getGenericStructDef(name string) (*hir.GenericStructDef, bool) {
	def, ok := self.getLocalGenericStructDef(name)
	if ok {
		return def, true
	}
	for iter := self.links.Iterator(); iter.Next(); {
		def, ok := iter.Value().getLocalGenericStructDef(name)
		if ok && def.GetPublic() {
			return def, true
		}
	}
	return nil, false
}

func (self *_PkgScope) GetGenericStructDef(pkg, name string) (*hir.GenericStructDef, bool) {
	if pkg == "" {
		return self.getGenericStructDef(name)
	}
	pkgScope := self.externs.Get(pkg)
	if pkgScope == nil {
		return nil, false
	}
	def, ok := pkgScope.getLocalGenericStructDef(name)
	if !ok || !def.GetPublic() {
		return nil, false
	}
	return def, true
}

// 本地作用域
type _LocalScope interface {
	_Scope
	GetParent() _Scope
	GetFuncScope() *_FuncScope
	GetPkgScope() *_PkgScope
	GetFunc() hir.GlobalFunc
	SetLoop(loop hir.Loop)
	GetLoop() hir.Loop
	GetStructScope()hir.GlobalStruct
	IsInStructScope(st *hir.StructType)bool
}

// 函数作用域
type _FuncScope struct {
	_BlockScope
	parent *_PkgScope
	def    hir.GlobalFunc
}

func _NewFuncScope(p *_PkgScope, def hir.GlobalFunc) *_FuncScope {
	self := &_FuncScope{
		parent: p,
		def:    def,
	}
	self._BlockScope = *_NewBlockScope(self)
	return self
}

func (self *_FuncScope) SetValue(name string, v hir.Ident) bool {
	return self._BlockScope.SetValue(name, v)
}

func (self *_FuncScope) GetValue(pkg, name string) (hir.Ident, bool) {
	if pkg != "" {
		return self.parent.GetValue(pkg, name)
	}
	if self.values.ContainKey(name) {
		return self.values.Get(name), true
	}
	return self.parent.GetValue("", name)
}

func (self *_FuncScope) GetParent() _Scope {
	return self.parent
}

func (self *_FuncScope) GetFuncScope() *_FuncScope {
	return self
}

func (self *_FuncScope) GetPkgScope() *_PkgScope {
	return self.parent
}

func (self *_FuncScope) GetFunc() hir.GlobalFunc {
	return self.def
}

func (self *_FuncScope) GetStructScope()hir.GlobalStruct{
	switch f := self.GetFunc().(type) {
	case *hir.MethodDef:
		return f.Scope
	case *hir.GenericMethodDef:
		return f.Scope
	case *hir.GenericStructMethodDef:
		return f.Scope
	default:
		return nil
	}
}

func (self *_FuncScope) IsInStructScope(st *hir.StructType)bool{
	stScope := self.GetStructScope()
	if stScope == nil{
		return false
	}
	stName := stlbasic.TernaryAction(!strings.Contains(st.GetName(), "::"), func() string {
		return st.GetName()
	}, func() string {
		return strings.Split(st.GetName(), "::")[0]
	})
	return stScope.GetPackage().Equal(st.GetPackage()) && stScope.GetName() == stName
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

func (self *_BlockScope) GetFunc() hir.GlobalFunc {
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

func (self *_BlockScope) GetStructScope()hir.GlobalStruct{
	return self.GetFuncScope().GetStructScope()
}

func (self *_BlockScope) IsInStructScope(st *hir.StructType)bool{
	return self.GetFuncScope().IsInStructScope(st)
}
