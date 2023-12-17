package hir

import (
	"strings"

	"github.com/kkkunny/stl/container/hashmap"
	"github.com/kkkunny/stl/container/linkedhashmap"
	"github.com/kkkunny/stl/container/pair"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/util"
)

// Global 全局
type Global interface {
	GetPackage()Package
	GetPublic() bool
}

// StructDef 结构体定义
type StructDef struct {
	Pkg Package
	Public bool
	Name   string
	Fields linkedhashmap.LinkedHashMap[string, pair.Pair[bool, Type]]
	Methods hashmap.HashMap[string, *MethodDef]
}

func (self *StructDef) GetPackage()Package{
	return self.Pkg
}

func (self *StructDef) GetPublic() bool {
	return self.Public
}

func (self *StructDef) Impl(t *TraitDef)bool{
	for targetIter:=t.Methods.Iterator(); targetIter.Next(); {
		target := targetIter.Value()
		if self.GetImplMethod(target.First, target.Second) == nil{
			return false
		}
	}
	return true
}

func (self *StructDef) GetImplMethod(name string, ft *FuncType)*MethodDef{
	for iter:=self.Methods.Values().Iterator(); iter.Next(); {
		fun := iter.Value()
		if fun.Name == name && fun.GetMethodType().Equal(ft){
			return fun
		}
	}
	return nil
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
}

func (self *MethodDef) GetPackage()Package{
	return self.Pkg
}

func (self *MethodDef) GetPublic() bool {
	return self.Public
}

func (*MethodDef) stmt() {}

func (self *MethodDef) GetFuncType() *FuncType {
	params := lo.Map(self.Params, func(item *Param, index int) Type {
		return item.GetType()
	})
	return &FuncType{
		Ret:    self.Ret,
		Params: append([]Type{self.Scope}, params...),
	}
}

func (self *MethodDef) GetType() Type {
	return self.GetFuncType()
}

func (self *MethodDef) Mutable() bool {
	return false
}

func (self *MethodDef) GetMethodType() *FuncType {
	params := lo.Map(self.Params, func(item *Param, index int) Type {
		return item.GetType()
	})
	return &FuncType{
		Ret:    self.Ret,
		Params: params,
	}
}

// GenericFuncDef 泛型函数定义
type GenericFuncDef struct {
	Pkg Package
	Public     bool
	Name       string
	GenericParams linkedhashmap.LinkedHashMap[string, *GenericParam]
	Params     []*Param
	Ret        Type
	Body       *Block

	Instances hashmap.HashMap[string, *GenericFuncInstance]
}

func (self *GenericFuncDef) GetPackage()Package{
	return self.Pkg
}

func (self *GenericFuncDef) GetPublic() bool {
	return self.Public
}

func (self *GenericFuncDef) GetFuncType() *FuncType {
	params := lo.Map(self.Params, func(item *Param, index int) Type {
		return item.GetType()
	})
	return &FuncType{
		Ret:    self.Ret,
		Params: params,
	}
}

func (self *GenericFuncDef) AddInstance(genericArg ...Type)*GenericFuncInstance{
	if uint(len(genericArg)) != self.GenericParams.Length(){
		panic("unreachable")
	}

	typeNames := lo.Map(genericArg, func(item Type, _ int) string {
		return item.String()
	})
	key := strings.Join(typeNames, ", ")

	inst := self.Instances.Get(key)
	if inst != nil{
		return inst
	}

	inst = &GenericFuncInstance{
		Define: self,
		Params: genericArg,
	}
	self.Instances.Set(key, inst)
	return inst
}

// GenericStructDef 泛型结构体定义
type GenericStructDef struct {
	Pkg Package
	Public bool
	Name   string
	GenericParams linkedhashmap.LinkedHashMap[string, *GenericParam]
	Fields linkedhashmap.LinkedHashMap[string, pair.Pair[bool, Type]]

	Instances hashmap.HashMap[string, *GenericStructInstance]
}

func (self *GenericStructDef) GetPackage()Package{
	return self.Pkg
}

func (self *GenericStructDef) GetPublic() bool {
	return self.Public
}

func (self *GenericStructDef) AddInstance(genericArg ...Type)*GenericStructInstance{
	if uint(len(genericArg)) != self.GenericParams.Length(){
		panic("unreachable")
	}

	typeNames := lo.Map(genericArg, func(item Type, _ int) string {
		return item.String()
	})
	key := strings.Join(typeNames, ", ")

	inst := self.Instances.Get(key)
	if inst != nil{
		return inst
	}

	inst = &GenericStructInstance{
		Define: self,
		Params: genericArg,
	}
	self.Instances.Set(key, inst)
	return inst
}

// TraitDef 特性定义
type TraitDef struct {
	Pkg Package
	Name string
	Methods hashmap.HashMap[string, *FuncType]
}

func (self *TraitDef) GetPackage()Package{
	return self.Pkg
}

func DefaultTrait(t Type)*TraitDef {
	return &TraitDef{Methods: hashmap.NewHashMapWith[string, *FuncType]("default", &FuncType{Ret: t})}
}
