package analyse

import (
	"github.com/kkkunny/stl/container/hashmap"
	"github.com/kkkunny/stl/container/linkedhashset"

	. "github.com/kkkunny/Sim/mean"
	"github.com/kkkunny/Sim/util"
)

// 作用域
type _Scope interface {
	SetValue(name string, v Ident) bool
	GetValue(pkg, name string) (Ident, bool)
}

// 包作用域
type _PkgScope struct {
	path    string
	externs hashmap.HashMap[string, *_PkgScope]
	links   linkedhashset.LinkedHashSet[*_PkgScope]
	values  hashmap.HashMap[string, Ident]
	structs hashmap.HashMap[string, *StructType]
}

func _NewPkgScope(path string) *_PkgScope {
	return &_PkgScope{
		path:    path,
		externs: hashmap.NewHashMap[string, *_PkgScope](),
		links:   linkedhashset.NewLinkedHashSet[*_PkgScope](),
		values:  hashmap.NewHashMap[string, Ident](),
		structs: hashmap.NewHashMap[string, *StructType](),
	}
}

// IsBuildIn 是否是buildin包
func (self *_PkgScope) IsBuildIn() bool {
	return self.path == util.GetBuildInPackagePath()
}

func (self *_PkgScope) SetValue(name string, v Ident) bool {
	if self.values.ContainKey(name) {
		return false
	}
	self.values.Set(name, v)
	return true
}

func (self *_PkgScope) getLocalValue(name string) (Ident, bool) {
	return self.values.Get(name), self.values.ContainKey(name)
}

func (self *_PkgScope) getValue(name string) (Ident, bool) {
	v, ok := self.getLocalValue(name)
	if ok {
		return v, true
	}
	for iter := self.links.Iterator(); iter.Next(); {
		v, ok := iter.Value().getLocalValue(name)
		if ok && v.(Global).GetPublic() {
			return v, true
		}
	}
	return nil, false
}

func (self *_PkgScope) GetValue(pkg, name string) (Ident, bool) {
	if pkg == "" {
		return self.getValue(name)
	}
	pkgScope := self.externs.Get(pkg)
	if pkgScope == nil {
		return nil, false
	}
	v, ok := pkgScope.GetValue("", name)
	if !ok || !v.(Global).GetPublic() {
		return nil, false
	}
	return v, true
}

func (self *_PkgScope) SetStruct(st *StructType) bool {
	if self.structs.ContainKey(st.Name) {
		return false
	}
	self.structs.Set(st.Name, st)
	return true
}

func (self *_PkgScope) getLocalStruct(name string) (*StructType, bool) {
	return self.structs.Get(name), self.structs.ContainKey(name)
}

func (self *_PkgScope) getStruct(name string) (*StructType, bool) {
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

func (self *_PkgScope) GetStruct(pkg, name string) (*StructType, bool) {
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
	GetRetType() Type
	SetLoop(loop Loop)
	GetLoop() Loop
}

// 函数作用域
type _FuncScope struct {
	_BlockScope
	parent  *_PkgScope
	retType Type
}

func _NewFuncScope(p *_PkgScope, ret Type) *_FuncScope {
	self := &_FuncScope{
		parent:  p,
		retType: ret,
	}
	self._BlockScope = *_NewBlockScope(self)
	return self
}

func (self *_FuncScope) SetValue(name string, v Ident) bool {
	return self._BlockScope.SetValue(name, v)
}

func (self *_FuncScope) GetValue(pkg, name string) (Ident, bool) {
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

func (self *_FuncScope) GetRetType() Type {
	return self.retType
}

// 代码块作用域
type _BlockScope struct {
	parent _LocalScope
	values hashmap.HashMap[string, Ident]
	loop   Loop
}

func _NewBlockScope(p _LocalScope) *_BlockScope {
	return &_BlockScope{
		parent: p,
		values: hashmap.NewHashMap[string, Ident](),
	}
}

func (self *_BlockScope) SetValue(name string, v Ident) bool {
	self.values.Set(name, v)
	return true
}

func (self *_BlockScope) GetValue(pkg, name string) (Ident, bool) {
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

func (self *_BlockScope) GetRetType() Type {
	return self.parent.GetRetType()
}

func (self *_BlockScope) SetLoop(loop Loop) {
	self.loop = loop
}

func (self *_BlockScope) GetLoop() Loop {
	if self.loop != nil {
		return self.loop
	}
	if self.parent != nil {
		return self.parent.GetLoop()
	}
	return nil
}
