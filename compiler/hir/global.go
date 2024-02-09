package hir

import (
	"github.com/kkkunny/stl/container/hashmap"
	stlslices "github.com/kkkunny/stl/slices"

	"github.com/kkkunny/Sim/util"
)

// Global 全局
type Global interface {
	GetPackage() Package
	GetPublic() bool
}

// GlobalVarDef 全局变量定义
type GlobalVarDef struct {
	VarDecl
	Pkg        Package
	Public     bool
	ExternName string
	Value      util.Option[Expr]
}

func (self *GlobalVarDef) GetPackage() Package {
	return self.Pkg
}

func (self *GlobalVarDef) GetPublic() bool {
	return self.Public
}

func (*GlobalVarDef) stmt() {}

func (*GlobalVarDef) ident() {}

// MultiGlobalVarDef 多全局变量定义
type MultiGlobalVarDef struct {
	Vars  []*GlobalVarDef
	Value Expr
}

func (self *MultiGlobalVarDef) GetPackage() Package {
	return self.Vars[0].GetPackage()
}

func (self *MultiGlobalVarDef) GetPublic() bool {
	for _, v := range self.Vars {
		if !v.Public {
			return false
		}
	}
	return true
}

func (*MultiGlobalVarDef) stmt() {}

type GlobalFuncOrMethod interface {
	Global
	GetFuncType() *FuncType
}

// FuncDef 函数定义
type FuncDef struct {
	Pkg        Package
	Public     bool
	ExternName string
	Name       string
	Params     []*Param
	Ret        Type
	Body       util.Option[*Block]

	InlineControl util.Option[bool]
	VarArg bool
}

func (self *FuncDef) GetPackage() Package {
	return self.Pkg
}

func (self *FuncDef) GetPublic() bool {
	return self.Public
}

func (*FuncDef) stmt() {}

func (self *FuncDef) GetFuncType() *FuncType {
	params := stlslices.Map(self.Params, func(_ int, e *Param) Type {
		return e.GetType()
	})
	return NewFuncType(self.Ret, params...)
}

func (self *FuncDef) GetType() Type {
	return self.GetFuncType()
}

func (self *FuncDef) Mutable() bool {
	return false
}

func (*FuncDef) ident() {}

func (*FuncDef) globalFunc() {}

type GlobalMethod interface {
	GlobalFuncOrMethod
	GetSelfType() GlobalType
	GetMethodType() *FuncType
	IsStatic() bool
}

// MethodDef 方法定义
type MethodDef struct {
	Scope *TypeDef
	FuncDef
}

func (self *MethodDef) GetSelfType() GlobalType {
	return self.Scope
}

func (self *MethodDef) GetMethodType() *FuncType {
	ft := self.GetFuncType()
	if !self.IsStatic() {
		ft.Params = ft.Params[1:]
	}
	return ft
}

func (self *MethodDef) IsStatic() bool {
	if len(self.Params) == 0 {
		return true
	}
	firstParam := self.Params[0]
	return firstParam.Name != "self" ||
		(!NewRefType(false, self.GetSelfType()).EqualTo(firstParam.GetType()) &&
		!NewRefType(true, self.GetSelfType()).EqualTo(firstParam.GetType()))
}

// GlobalType 类型定义
type GlobalType interface {
	Global
	Type
	GetName() string
}

// TypeDef 类型定义
type TypeDef struct {
	Pkg     Package
	Public  bool
	Name    string
	Target Type
	Methods hashmap.HashMap[string, GlobalMethod]
}

func (self *TypeDef) GetPackage() Package {
	return self.Pkg
}

func (self *TypeDef) GetPublic() bool {
	return self.Public
}

func (self *TypeDef) GetName() string {
	return self.Name
}

// TypeAliasDef 类型别名定义
type TypeAliasDef struct {
	Pkg    Package
	Public bool
	Name   string
	Target Type
}

func (self *TypeAliasDef) GetPackage() Package {
	return self.Pkg
}

func (self *TypeAliasDef) GetPublic() bool {
	return self.Public
}

func (self *TypeAliasDef) GetName() string {
	return self.Name
}
