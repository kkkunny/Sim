package hir

import (
	"github.com/kkkunny/stl/container/hashmap"
	"github.com/kkkunny/stl/container/linkedhashmap"
	stlslices "github.com/kkkunny/stl/slices"
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

	genericParams []Type // 泛型结构体实例化时使用，为泛型结构体方法实例化提供泛型参数
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
	Value      Expr
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
	Global
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
	return !self.GetSelfType().EqualTo(self.Params[0].GetType())
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

type GlobalGenericDef interface {
	Global
	GetGenericParams() linkedhashmap.LinkedHashMap[string, *GenericIdentType]
}

// GenericFuncDef 泛型函数定义
type GenericFuncDef struct {
	Pkg           Package
	Public        bool
	Name          string
	GenericParams linkedhashmap.LinkedHashMap[string, *GenericIdentType]
	Params        []*Param
	Ret           Type
	Body          *Block

	NoReturn      bool
	InlineControl util.Option[bool]
}

func (self *GenericFuncDef) GetPackage() Package {
	return self.Pkg
}

func (self *GenericFuncDef) GetPublic() bool {
	return self.Public
}

func (self *GenericFuncDef) GetFuncType() *FuncType {
	return &FuncType{
		Ret: self.Ret,
		Params: lo.Map(self.Params, func(item *Param, index int) Type {
			return item.GetType()
		}),
	}
}

func (self *GenericFuncDef) GetGenericParams() linkedhashmap.LinkedHashMap[string, *GenericIdentType] {
	return self.GenericParams
}

func (*GenericFuncDef) globalFunc() {}

// GenericStructDef 泛型结构体定义
type GenericStructDef struct {
	Pkg           Package
	Public        bool
	Name          string
	GenericParams linkedhashmap.LinkedHashMap[string, *GenericIdentType]
	Fields        linkedhashmap.LinkedHashMap[string, Field]
	Methods       hashmap.HashMap[string, *GenericStructMethodDef]
}

func (self *GenericStructDef) GetPackage() Package {
	return self.Pkg
}

func (self *GenericStructDef) GetPublic() bool {
	return self.Public
}

func (self *GenericStructDef) GetName() string {
	return self.Name
}

func (self *GenericStructDef) GetGenericParams() linkedhashmap.LinkedHashMap[string, *GenericIdentType] {
	return self.GenericParams
}

func (*GenericStructDef) globalStruct() {}

// GenericStructMethodDef 泛型结构体方法定义
type GenericStructMethodDef struct {
	Scope *GenericStructDef
	GenericFuncDef
}

func (self *GenericStructMethodDef) GetSelfType() TypeDef {
	return &GenericStructInst{
		Define: self.Scope,
		Params: stlslices.Map(self.Scope.GenericParams.Values().ToSlice(), func(_ int, e *GenericIdentType) Type {
			return e
		}),
	}
}

func (self *GenericStructMethodDef) GetMethodType() *FuncType {
	ft := self.GetFuncType()
	if !self.IsStatic() {
		ft.Params = ft.Params[1:]
	}
	return ft
}

func (self *GenericStructMethodDef) GetGenericParams() (res linkedhashmap.LinkedHashMap[string, *GenericIdentType]) {
	for iter := self.Scope.GetGenericParams().Iterator(); iter.Next(); {
		res.Set(iter.Value().First, iter.Value().Second)
	}
	for iter := self.GenericParams.Iterator(); iter.Next(); {
		res.Set(iter.Value().First, iter.Value().Second)
	}
	return res
}

func (self *GenericStructMethodDef) IsStatic() bool {
	if len(self.Params) == 0 {
		return true
	}
	return !self.GetSelfType().EqualTo(self.Params[0].GetType())
}

// GenericMethodDef 泛型方法定义
type GenericMethodDef struct {
	Scope *StructDef
	GenericFuncDef
}

func (self *GenericMethodDef) GetSelfType() TypeDef {
	return self.Scope
}

func (self *GenericMethodDef) GetMethodType() *FuncType {
	ft := self.GetFuncType()
	if !self.IsStatic() {
		ft.Params = ft.Params[1:]
	}
	return ft
}

func (self *GenericMethodDef) GetGenericParams() linkedhashmap.LinkedHashMap[string, *GenericIdentType] {
	return self.GenericParams
}

func (self *GenericMethodDef) IsStatic() bool {
	if len(self.Params) == 0 {
		return true
	}
	return !self.GetSelfType().EqualTo(self.Params[0].GetType())
}

// TraitDef 特性定义
type TraitDef struct {
	Pkg     Package
	Public  bool
	Name    string
	Methods hashmap.HashMap[string, *FuncType]
}

func (self *TraitDef) GetPackage() Package {
	return self.Pkg
}

func (self *TraitDef) GetPublic() bool {
	return self.Public
}
