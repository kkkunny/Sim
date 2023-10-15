package analyse

import (
	"github.com/kkkunny/stl/container/hashmap"

	. "github.com/kkkunny/Sim/mean"
)

// 作用域
type _Scope interface {
	SetValue(name string, v Expr) bool
	GetValue(name string) (Expr, bool)
}

// 包作用域
type _PkgScope struct {
	values hashmap.HashMap[string, Expr]
}

func _NewPkgScope() *_PkgScope {
	return &_PkgScope{values: hashmap.NewHashMap[string, Expr]()}
}

func (self *_PkgScope) SetValue(name string, v Expr) bool {
	if self.values.ContainKey(name) {
		return false
	}
	self.values.Set(name, v)
	return true
}

func (self *_PkgScope) GetValue(name string) (Expr, bool) {
	return self.values.Get(name), self.values.ContainKey(name)
}

// 本地作用域
type _LocalScope interface {
	_Scope
	GetParent() _Scope
	GetFuncScope() *_FuncScope
	GetPkgScope() *_PkgScope
	GetRetType() Type
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

func (self *_FuncScope) SetValue(name string, v Expr) bool {
	return self._BlockScope.SetValue(name, v)
}

func (self *_FuncScope) GetValue(name string) (Expr, bool) {
	if self.values.ContainKey(name) {
		self.values.Get(name)
	}
	return self.parent.GetValue(name)
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
	values hashmap.HashMap[string, Expr]
}

func _NewBlockScope(p _LocalScope) *_BlockScope {
	return &_BlockScope{
		parent: p,
		values: hashmap.NewHashMap[string, Expr](),
	}
}

func (self *_BlockScope) SetValue(name string, v Expr) bool {
	self.values.Set(name, v)
	return true
}

func (self *_BlockScope) GetValue(name string) (Expr, bool) {
	if self.values.ContainKey(name) {
		self.values.Get(name)
	}
	return self.parent.GetValue(name)
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
