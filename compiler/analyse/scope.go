package analyse

import (
	"github.com/kkkunny/stl/container/hashmap"
	"github.com/kkkunny/stl/container/linkedhashset"

	"github.com/kkkunny/Sim/mean"

	"github.com/kkkunny/Sim/util"
)

// 作用域
type _Scope interface {
	SetValue(name string, v mean.Ident) bool
	GetValue(pkg, name string) (mean.Ident, bool)
}

// 包作用域
type _PkgScope struct {
	path    string
	externs hashmap.HashMap[string, *_PkgScope]
	links   linkedhashset.LinkedHashSet[*_PkgScope]
	values  hashmap.HashMap[string, mean.Ident]
	structs hashmap.HashMap[string, *mean.StructType]
}

func _NewPkgScope(path string) *_PkgScope {
	return &_PkgScope{
		path:    path,
		externs: hashmap.NewHashMap[string, *_PkgScope](),
		links:   linkedhashset.NewLinkedHashSet[*_PkgScope](),
		values:  hashmap.NewHashMap[string, mean.Ident](),
		structs: hashmap.NewHashMap[string, *mean.StructType](),
	}
}

// IsBuildIn 是否是buildin包
func (self *_PkgScope) IsBuildIn() bool {
	return self.path == util.GetBuildInPackagePath()
}

func (self *_PkgScope) SetValue(name string, v mean.Ident) bool {
	if self.values.ContainKey(name) {
		return false
	}
	self.values.Set(name, v)
	return true
}

func (self *_PkgScope) getLocalValue(name string) (mean.Ident, bool) {
	return self.values.Get(name), self.values.ContainKey(name)
}

func (self *_PkgScope) getValue(name string) (mean.Ident, bool) {
	v, ok := self.getLocalValue(name)
	if ok {
		return v, true
	}
	for iter := self.links.Iterator(); iter.Next(); {
		v, ok := iter.Value().getLocalValue(name)
		if ok && v.(mean.Global).GetPublic() {
			return v, true
		}
	}
	return nil, false
}

func (self *_PkgScope) GetValue(pkg, name string) (mean.Ident, bool) {
	if pkg == "" {
		return self.getValue(name)
	}
	pkgScope := self.externs.Get(pkg)
	if pkgScope == nil {
		return nil, false
	}
	v, ok := pkgScope.GetValue("", name)
	if !ok || !v.(mean.Global).GetPublic() {
		return nil, false
	}
	return v, true
}

func (self *_PkgScope) SetStruct(st *mean.StructType) bool {
	if self.structs.ContainKey(st.Name) {
		return false
	}
	self.structs.Set(st.Name, st)
	return true
}

func (self *_PkgScope) getLocalStruct(name string) (*mean.StructType, bool) {
	return self.structs.Get(name), self.structs.ContainKey(name)
}

func (self *_PkgScope) getStruct(name string) (*mean.StructType, bool) {
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

func (self *_PkgScope) GetStruct(pkg, name string) (*mean.StructType, bool) {
	if pkg == "" {
		return self.getStruct(name)
	}
	pkgScope := self.externs.Get(pkg)
	if pkgScope == nil {
		return nil, false
	}
	t, ok := pkgScope.GetStruct("", name)
	if !ok || !t.GetPublic() {
		return nil, false
	}
	return t, true
}

// 本地作用域
type _LocalScope interface {
	_Scope
	GetParent() _Scope
	GetFuncScope() *_FuncScope
	GetPkgScope() *_PkgScope
	GetFunc() mean.Function
	SetLoop(loop mean.Loop)
	GetLoop() mean.Loop
}

// 函数作用域
type _FuncScope struct {
	_BlockScope
	parent *_PkgScope
	def    mean.Function
}

func _NewFuncScope(p *_PkgScope, def mean.Function) *_FuncScope {
	self := &_FuncScope{
		parent: p,
		def:    def,
	}
	self._BlockScope = *_NewBlockScope(self)
	return self
}

func (self *_FuncScope) SetValue(name string, v mean.Ident) bool {
	return self._BlockScope.SetValue(name, v)
}

func (self *_FuncScope) GetValue(pkg, name string) (mean.Ident, bool) {
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

func (self *_FuncScope) GetFunc() mean.Function {
	return self.def
}

// 代码块作用域
type _BlockScope struct {
	parent _LocalScope
	values hashmap.HashMap[string, mean.Ident]
	loop   mean.Loop
}

func _NewBlockScope(p _LocalScope) *_BlockScope {
	return &_BlockScope{
		parent: p,
		values: hashmap.NewHashMap[string, mean.Ident](),
	}
}

func (self *_BlockScope) SetValue(name string, v mean.Ident) bool {
	self.values.Set(name, v)
	return true
}

func (self *_BlockScope) GetValue(pkg, name string) (mean.Ident, bool) {
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

func (self *_BlockScope) GetFunc() mean.Function {
	return self.parent.GetFunc()
}

func (self *_BlockScope) SetLoop(loop mean.Loop) {
	self.loop = loop
}

func (self *_BlockScope) GetLoop() mean.Loop {
	if self.loop != nil {
		return self.loop
	}
	if self.parent != nil {
		return self.parent.GetLoop()
	}
	return nil
}
