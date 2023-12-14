package analyse

import (
	"github.com/kkkunny/stl/container/either"
	"github.com/kkkunny/stl/container/hashmap"
	"github.com/kkkunny/stl/container/linkedhashset"
	"github.com/kkkunny/stl/container/pair"

	"github.com/kkkunny/Sim/ast"
	"github.com/kkkunny/Sim/hir"

	"github.com/kkkunny/Sim/util"
)

// 作用域
type _Scope interface {
	SetValue(name string, v hir.Ident) bool
	GetValue(pkg, name string) (hir.Ident, bool)
}

// 包作用域
type _PkgScope struct {
	path      string
	externs   hashmap.HashMap[string, *_PkgScope]
	links     linkedhashset.LinkedHashSet[*_PkgScope]

	values    hashmap.HashMap[string, hir.Ident]
	structs   hashmap.HashMap[string, *hir.StructType]
	typeAlias hashmap.HashMap[string, pair.Pair[bool, either.Either[*ast.TypeAlias, hir.Type]]]
	traits hashmap.HashMap[string, pair.Pair[bool, *ast.Trait]]

	genericFunctions hashmap.HashMap[string, *hir.GenericFuncDef]
}

func _NewPkgScope(path string) *_PkgScope {
	return &_PkgScope{
		path:      path,
		externs:   hashmap.NewHashMap[string, *_PkgScope](),
		links:     linkedhashset.NewLinkedHashSet[*_PkgScope](),
		values:    hashmap.NewHashMap[string, hir.Ident](),
		structs:   hashmap.NewHashMap[string, *hir.StructType](),
		typeAlias: hashmap.NewHashMap[string, pair.Pair[bool, either.Either[*ast.TypeAlias, hir.Type]]](),
		traits: hashmap.NewHashMap[string, pair.Pair[bool, *ast.Trait]](),
		genericFunctions: hashmap.NewHashMap[string, *hir.GenericFuncDef](),
	}
}

// IsBuildIn 是否是buildin包
func (self *_PkgScope) IsBuildIn() bool {
	return self.path == util.GetBuildInPackagePath()
}

func (self *_PkgScope) SetValue(name string, v hir.Ident) bool {
	if _, ok := self.getValue(name); ok {
		return false
	}
	if _, ok := self.getGenericFunction(name); ok {
		return false
	}
	self.values.Set(name, v)
	return true
}

func (self *_PkgScope) getLocalValue(name string) (hir.Ident, bool) {
	return self.values.Get(name), self.values.ContainKey(name)
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

func (self *_PkgScope) SetStruct(st *hir.StructType) bool {
	if _, ok := self.getStruct(st.Name); ok {
		return false
	}
	if _, ok := self.getTypeAlias(st.Name); ok {
		return false
	}
	self.structs.Set(st.Name, st)
	return true
}

func (self *_PkgScope) getLocalStruct(name string) (*hir.StructType, bool) {
	return self.structs.Get(name), self.structs.ContainKey(name)
}

func (self *_PkgScope) getStruct(name string) (*hir.StructType, bool) {
	st, ok := self.getLocalStruct(name)
	if ok {
		return st, true
	}
	for iter := self.links.Iterator(); iter.Next(); {
		st, ok := iter.Value().getLocalStruct(name)
		if ok && st.GetPublic() {
			return st, true
		}
	}
	return nil, false
}

func (self *_PkgScope) GetStruct(pkg, name string) (*hir.StructType, bool) {
	if pkg == "" {
		return self.getStruct(name)
	}
	pkgScope := self.externs.Get(pkg)
	if pkgScope == nil {
		return nil, false
	}
	t, ok := pkgScope.getLocalStruct(name)
	if !ok || !t.GetPublic() {
		return nil, false
	}
	return t, true
}

func (self *_PkgScope) SetTrait(pub bool, node *ast.Trait) bool {
	if self.traits.ContainKey(node.Name.Source()) {
		return false
	}
	self.traits.Set(node.Name.Source(), pair.NewPair(pub, node))
	return true
}

func (self *_PkgScope) getLocalTrait(name string) (pair.Pair[bool, *ast.Trait], bool) {
	return self.traits.Get(name), self.traits.ContainKey(name)
}

func (self *_PkgScope) getTrait(name string) (*ast.Trait, bool) {
	info, ok := self.getLocalTrait(name)
	if ok {
		return info.Second, true
	}
	for iter := self.links.Iterator(); iter.Next(); {
		info, ok := iter.Value().getLocalTrait(name)
		if ok && info.First {
			return info.Second, true
		}
	}
	return nil, false
}

func (self *_PkgScope) GetTrait(pkg, name string) (*ast.Trait, bool) {
	if pkg == "" {
		return self.getTrait(name)
	}
	pkgScope := self.externs.Get(pkg)
	if pkgScope == nil {
		return nil, false
	}
	info, ok := pkgScope.getLocalTrait(name)
	if !ok || !info.First {
		return nil, false
	}
	return info.Second, true
}

func (self *_PkgScope) DeclTypeAlias(name string, node *ast.TypeAlias) bool {
	if _, ok := self.getStruct(name); ok {
		return false
	}
	if _, ok := self.getTypeAlias(name); ok {
		return false
	}
	self.typeAlias.Set(name, pair.NewPair(node.Public, either.Left[*ast.TypeAlias, hir.Type](node)))
	return true
}

func (self *_PkgScope) DefTypeAlias(name string, t hir.Type) {
	if !self.typeAlias.ContainKey(name){
		panic("unreachable")
	}
	self.typeAlias.Set(name, pair.NewPair(self.typeAlias.Get(name).First, either.Right[*ast.TypeAlias, hir.Type](t)))
}

func (self *_PkgScope) getLocalTypeAlias(name string) (pair.Pair[bool, either.Either[*ast.TypeAlias, hir.Type]], bool) {
	return self.typeAlias.Get(name), self.typeAlias.ContainKey(name)
}

func (self *_PkgScope) getTypeAlias(name string) (either.Either[*ast.TypeAlias, hir.Type], bool) {
	info, ok := self.getLocalTypeAlias(name)
	if ok {
		return info.Second, true
	}
	for iter := self.links.Iterator(); iter.Next(); {
		info, ok = iter.Value().getLocalTypeAlias(name)
		if ok && info.First {
			return info.Second, true
		}
	}
	var tmp either.Either[*ast.TypeAlias, hir.Type]
	return tmp, false
}

func (self *_PkgScope) GetTypeAlias(pkg, name string) (either.Either[*ast.TypeAlias, hir.Type], bool) {
	if pkg == "" {
		return self.getTypeAlias(name)
	}
	pkgScope := self.externs.Get(pkg)
	if pkgScope == nil {
		var tmp either.Either[*ast.TypeAlias, hir.Type]
		return tmp, false
	}
	data, ok := pkgScope.getLocalTypeAlias(name)
	return data.Second, ok && data.First
}

func (self *_PkgScope) SetGenericFunction(name string, f *hir.GenericFuncDef) bool {
	if _, ok := self.getValue(name); ok {
		return false
	}
	if _, ok := self.getGenericFunction(name); ok {
		return false
	}
	self.genericFunctions.Set(name, f)
	return true
}

func (self *_PkgScope) getLocalGenericFunction(name string) (*hir.GenericFuncDef, bool) {
	return self.genericFunctions.Get(name), self.genericFunctions.ContainKey(name)
}

func (self *_PkgScope) getGenericFunction(name string) (*hir.GenericFuncDef, bool) {
	f, ok := self.getLocalGenericFunction(name)
	if ok {
		return f, true
	}
	for iter := self.links.Iterator(); iter.Next(); {
		f, ok := iter.Value().getLocalGenericFunction(name)
		if ok && f.GetPublic() {
			return f, true
		}
	}
	return nil, false
}

func (self *_PkgScope) GetGenericFunction(pkg, name string) (*hir.GenericFuncDef, bool) {
	if pkg == "" {
		return self.getGenericFunction(name)
	}
	pkgScope := self.externs.Get(pkg)
	if pkgScope == nil {
		return nil, false
	}
	f, ok := pkgScope.getLocalGenericFunction(name)
	if !ok || !f.GetPublic() {
		return nil, false
	}
	return f, true
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
