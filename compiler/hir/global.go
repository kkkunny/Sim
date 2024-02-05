package hir

import (
	"github.com/kkkunny/stl/container/hashmap"
	"github.com/kkkunny/stl/container/linkedhashmap"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/util"
)

// Global 全局
type Global interface {
	GetPackage() Package
	GetPublic() bool
}

// TypeDef 类型定义
type TypeDef interface {
	Global
	Type
	GetName() string
}

// GlobalStruct 全局结构体
type GlobalStruct interface {
	Global
	GetName() string
	globalStruct()
}

type Field struct {
	Public  bool
	Mutable bool
	Type    Type
}

// StructDef 结构体定义
type StructDef struct {
	Pkg     Package
	Public  bool
	Name    string
	Fields  linkedhashmap.LinkedHashMap[string, Field]
	Methods hashmap.HashMap[string, GlobalMethod]

	genericArgs []Type // 泛型结构体实例化时使用，为泛型结构体方法实例化提供泛型参数
}

func (self *StructDef) GetPackage() Package {
	return self.Pkg
}

func (self *StructDef) GetPublic() bool {
	return self.Public
}

func (self *StructDef) GetName() string {
	return self.Name
}

func (*StructDef) globalStruct() {}

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

type GlobalFunc interface {
	GlobalFuncOrMethod
	globalFunc()
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

	NoReturn      bool
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
	params := lo.Map(self.Params, func(item *Param, index int) Type {
		return item.GetType()
	})
	return &FuncType{
		Ret:    self.Ret,
		Params: params,
	}
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
	GetSelfType() TypeDef
	GetMethodType() *FuncType
	IsStatic() bool
}

// MethodDef 方法定义
type MethodDef struct {
	Scope *StructDef
	FuncDef
}

func (self *MethodDef) GetSelfType() TypeDef {
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
	selfRef := &RefType{Elem: self.GetSelfType()}
	return !selfRef.EqualTo(self.Params[0].GetType())
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
