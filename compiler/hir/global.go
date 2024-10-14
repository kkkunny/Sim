package hir

import (
	"github.com/kkkunny/stl/container/hashmap"
	stliter "github.com/kkkunny/stl/container/iter"
	"github.com/kkkunny/stl/container/linkedhashmap"
	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
	"github.com/kkkunny/stl/container/tuple"
	stlval "github.com/kkkunny/stl/value"
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
	Value      optional.Optional[Expr]
}

func (self *GlobalVarDef) GetPackage() Package {
	return self.Pkg
}

func (self *GlobalVarDef) GetPublic() bool {
	return self.Public
}

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

type GlobalFuncOrMethod interface {
	Global
	CallableDef
}

// FuncDef 函数定义
type FuncDef struct {
	Pkg        Package
	Public     bool
	ExternName string
	FuncDecl
	Body optional.Optional[*Block]

	InlineControl optional.Optional[bool]
	VarArg        bool
}

func (self *FuncDef) GetPackage() Package {
	return self.Pkg
}

func (self *FuncDef) GetPublic() bool {
	return self.Public
}

func (self *FuncDef) GetFuncType() *FuncType {
	return self.FuncDecl.GetType()
}

func (self *FuncDef) GetType() Type {
	return self.GetFuncType()
}

func (*FuncDef) Mutable() bool {
	return false
}

func (*FuncDef) globalFunc() {}

func (self *FuncDef) GetName() string {
	return self.Name
}

func (*FuncDef) Temporary() bool {
	return true
}

type GlobalMethod interface {
	GlobalFuncOrMethod
	GetSelfType() *TypeDef
	GetMethodType() *FuncType
	GetSelfParam() optional.Optional[*Param]
	IsStatic() bool
	IsRef() bool
}

// MethodDef 方法定义
type MethodDef struct {
	Scope *TypeDef
	FuncDef
}

func (self *MethodDef) GetSelfType() *TypeDef {
	return self.Scope
}

func (self *MethodDef) GetMethodType() *FuncType {
	ft := self.GetFuncType()
	if !self.IsStatic() {
		ft.Params = ft.Params[1:]
	}
	return ft
}

func (self *MethodDef) GetSelfParam() optional.Optional[*Param] {
	if len(self.Params) == 0 {
		return optional.None[*Param]()
	}
	selfType := self.GetSelfType()
	firstParam := self.Params[0]
	firstParamType := stlval.TernaryAction(IsType[*RefType](firstParam.GetType()), func() Type {
		return AsType[*RefType](firstParam.GetType()).Elem
	}, func() Type {
		return firstParam.GetType()
	})
	if !selfType.EqualTo(firstParamType) {
		return optional.None[*Param]()
	}
	return optional.Some(firstParam)
}

func (self *MethodDef) IsStatic() bool {
	return self.GetSelfParam().IsNone()
}

func (self *MethodDef) IsRef() bool {
	selfParam, ok := self.GetSelfParam().Value()
	if !ok {
		return false
	}
	return IsType[*RefType](selfParam.GetType())
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
	Target  Type
	Methods hashmap.HashMap[string, *MethodDef]
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

// Trait 特征
type Trait struct {
	Pkg     Package
	Public  bool
	Name    string
	Methods linkedhashmap.LinkedHashMap[string, *FuncDecl]
}

func (self *Trait) GetPackage() Package {
	return self.Pkg
}

func (self *Trait) GetPublic() bool {
	return self.Public
}

// HasBeImpled 类型是否实现了该特征
func (self *Trait) HasBeImpled(t Type) bool {
	ct, ok := TryCustomType(t)
	if !ok {
		return false
	}
	return stliter.All(self.Methods, func(e tuple.Tuple2[string, *FuncDecl]) bool {
		method := e.E2()
		impl := ct.Methods.Get(method.Name)
		if impl == nil {
			return false
		}
		return replaceAllSelfType(method.GetType(), ct).EqualTo(impl.GetFuncType())
	}) || self.HasBeImpled(ct.Target)
}

func (self *Trait) FirstMethodName() string {
	return stlslices.First(self.Methods.Values()).Name
}
