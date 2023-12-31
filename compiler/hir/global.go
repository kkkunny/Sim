package hir

import (
	"github.com/kkkunny/stl/container/hashmap"
	"github.com/kkkunny/stl/container/linkedhashmap"
	"github.com/kkkunny/stl/container/pair"
	stlslices "github.com/kkkunny/stl/slices"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/util"
)

// Global 全局
type Global interface {
	GetPackage()Package
	GetPublic() bool
}

// TypeDef 类型定义
type TypeDef interface {
	Global
	Type
	GetName()string
}

// StructDef 结构体定义
type StructDef struct {
	Pkg Package
	Public bool
	Name   string
	Fields linkedhashmap.LinkedHashMap[string, pair.Pair[bool, Type]]
	Methods hashmap.HashMap[string, GlobalMethod]

	genericParams []Type  // 泛型结构体实例化时使用，为泛型结构体方法实例化提供泛型参数
}

func (self *StructDef) GetPackage()Package{
	return self.Pkg
}

func (self *StructDef) GetPublic() bool {
	return self.Public
}

func (self *StructDef) GetName()string{
	return self.Name
}

// VarDef 变量定义
type VarDef struct {
	Pkg Package
	Public     bool
	Mut        bool
	Type       Type
	ExternName string
	Name       string
	Value      Expr
}

func (self *VarDef) GetPackage()Package{
	return self.Pkg
}

func (self *VarDef) GetPublic() bool {
	return self.Public
}

func (self *VarDef) GetName()string{
	return self.Name
}

func (*VarDef) stmt() {}

func (self *VarDef) GetType() Type {
	return self.Type
}

func (self *VarDef) Mutable() bool {
	return self.Mut
}

func (*VarDef) ident() {}

// MultiVarDef 多变量定义
type MultiVarDef struct {
	Vars []*VarDef
	Value      Expr
}

func (self *MultiVarDef) GetPackage()Package{
	return self.Vars[0].GetPackage()
}

func (self *MultiVarDef) GetPublic() bool {
	for _, v := range self.Vars{
		if !v.Public{
			return false
		}
	}
	return true
}

func (*MultiVarDef) stmt() {}

type GlobalFunc interface {
	Global
	GetFuncType()*FuncType
}

// FuncDef 函数定义
type FuncDef struct {
	Pkg Package
	Public     bool
	ExternName string
	Name       string
	Params     []*Param
	Ret        Type
	Body       util.Option[*Block]

	NoReturn bool
	InlineControl util.Option[bool]
}

func (self *FuncDef) GetPackage()Package{
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

type GlobalMethod interface {
	Global
	GetMethodType()*FuncType
}

// MethodDef 方法定义
type MethodDef struct {
	Pkg Package
	Public     bool
	Scope *StructDef
	Name       string
	SelfParam *Param
	Params     []*Param
	Ret        Type
	Body       *Block

	NoReturn bool
	InlineControl util.Option[bool]
}

func (self *MethodDef) GetPackage()Package{
	return self.Pkg
}

func (self *MethodDef) GetPublic() bool {
	return self.Public
}

func (*MethodDef) stmt() {}

func (self *MethodDef) GetFuncType() *FuncType {
	return &FuncType{
		Ret:    self.Ret,
		Params: append([]Type{self.Scope}, lo.Map(self.Params, func(item *Param, index int) Type {
			return item.GetType()
		})...),
	}
}

func (self *MethodDef) GetType() Type {
	return self.GetFuncType()
}

func (self *MethodDef) Mutable() bool {
	return false
}

func (self *MethodDef) GetMethodType() *FuncType {
	return &FuncType{
		Ret:    self.Ret,
		Params: lo.Map(self.Params, func(item *Param, index int) Type {
			return item.GetType()
		}),
	}
}

// TypeAliasDef 类型别名定义
type TypeAliasDef struct {
	Pkg Package
	Public bool
	Name   string
	Target Type
}

func (self *TypeAliasDef) GetPackage()Package{
	return self.Pkg
}

func (self *TypeAliasDef) GetPublic() bool {
	return self.Public
}

func (self *TypeAliasDef) GetName()string{
	return self.Name
}

type GlobalGenericDef interface {
	Global
	GetGenericParams()linkedhashmap.LinkedHashMap[string, *GenericIdentType]
}

// GenericFuncDef 泛型函数定义
type GenericFuncDef struct {
	Pkg Package
	Public     bool
	Name       string
	GenericParams linkedhashmap.LinkedHashMap[string, *GenericIdentType]
	Params     []*Param
	Ret        Type
	Body       *Block

	NoReturn bool
	InlineControl util.Option[bool]
}

func (self *GenericFuncDef) GetPackage()Package{
	return self.Pkg
}

func (self *GenericFuncDef) GetPublic() bool {
	return self.Public
}

func (self *GenericFuncDef) GetFuncType() *FuncType {
	return &FuncType{
		Ret:    self.Ret,
		Params: lo.Map(self.Params, func(item *Param, index int) Type {
			return item.GetType()
		}),
	}
}

func (self *GenericFuncDef) GetGenericParams()linkedhashmap.LinkedHashMap[string, *GenericIdentType]{
	return self.GenericParams
}

// GenericStructDef 泛型结构体定义
type GenericStructDef struct {
	Pkg Package
	Public bool
	Name   string
	GenericParams linkedhashmap.LinkedHashMap[string, *GenericIdentType]
	Fields linkedhashmap.LinkedHashMap[string, pair.Pair[bool, Type]]
	Methods hashmap.HashMap[string, *GenericStructMethodDef]
}

func (self *GenericStructDef) GetPackage()Package{
	return self.Pkg
}

func (self *GenericStructDef) GetPublic() bool {
	return self.Public
}

func (self *GenericStructDef) GetGenericParams()linkedhashmap.LinkedHashMap[string, *GenericIdentType]{
	return self.GenericParams
}

// GenericStructMethodDef 泛型结构体方法定义
type GenericStructMethodDef struct {
	Pkg Package
	Public     bool
	Scope *GenericStructDef
	Name       string
	SelfParam *Param
	Params     []*Param
	Ret        Type
	Body       *Block

	NoReturn bool
	InlineControl util.Option[bool]
}

func (self *GenericStructMethodDef) GetPackage()Package{
	return self.Pkg
}

func (self *GenericStructMethodDef) GetPublic() bool {
	return self.Public
}

func (self *GenericStructMethodDef) GetSelfType()*GenericStructInst{
	return &GenericStructInst{
		Define: self.Scope,
		Params: stlslices.Map(self.Scope.GenericParams.Values().ToSlice(), func(_ int, e *GenericIdentType) Type {
			return e
		}),
	}
}

func (self *GenericStructMethodDef) GetFuncType() *FuncType {
	return &FuncType{
		Ret:    self.Ret,
		Params: append([]Type{self.GetSelfType()}, lo.Map(self.Params, func(item *Param, index int) Type {
			return item.GetType()
		})...),
	}
}

func (self *GenericStructMethodDef) GetMethodType() *FuncType {
	return &FuncType{
		Ret:    self.Ret,
		Params: lo.Map(self.Params, func(item *Param, index int) Type {
			return item.GetType()
		}),
	}
}

func (self *GenericStructMethodDef) GetGenericParams()linkedhashmap.LinkedHashMap[string, *GenericIdentType]{
	return self.Scope.GetGenericParams()
}

// GenericMethodDef 泛型方法定义
type GenericMethodDef struct {
	Pkg Package
	Public     bool
	Scope *StructDef
	Name       string
	GenericParams linkedhashmap.LinkedHashMap[string, *GenericIdentType]
	SelfParam *Param
	Params     []*Param
	Ret        Type
	Body       *Block

	NoReturn bool
	InlineControl util.Option[bool]
}

func (self *GenericMethodDef) GetPackage()Package{
	return self.Pkg
}

func (self *GenericMethodDef) GetPublic() bool {
	return self.Public
}

func (self *GenericMethodDef) GetFuncType() *FuncType {
	return &FuncType{
		Ret:    self.Ret,
		Params: append([]Type{self.Scope}, lo.Map(self.Params, func(item *Param, index int) Type {
			return item.GetType()
		})...),
	}
}

func (self *GenericMethodDef) GetMethodType() *FuncType {
	return &FuncType{
		Ret:    self.Ret,
		Params: lo.Map(self.Params, func(item *Param, index int) Type {
			return item.GetType()
		}),
	}
}

func (self *GenericMethodDef) GetGenericParams()linkedhashmap.LinkedHashMap[string, *GenericIdentType]{
	return self.GenericParams
}
