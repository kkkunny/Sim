package mean

import (
	"github.com/kkkunny/stl/container/hashmap"
	"github.com/kkkunny/stl/container/linkedhashmap"
	"github.com/kkkunny/stl/container/pair"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/util"
)

// Global 全局
type Global interface {
	GetPublic() bool
}

// StructDef 结构体定义
type StructDef struct {
	Public bool
	Pkg string
	Name   string
	Fields linkedhashmap.LinkedHashMap[string, pair.Pair[bool, Type]]
	Methods hashmap.HashMap[string, *MethodDef]
}

func (self StructDef) GetPublic() bool {
	return self.Public
}

// Variable 变量定义
type Variable struct {
	Public     bool
	Mut        bool
	Type       Type
	ExternName string
	Name       string
	Value      Expr
}

func (self Variable) GetPublic() bool {
	return self.Public
}

func (*Variable) stmt() {}

func (self *Variable) GetType() Type {
	return self.Type
}

func (self *Variable) Mutable() bool {
	return self.Mut
}

func (*Variable) ident() {}

// Function 函数
type Function interface {
	Expr
	GetFuncType()*FuncType
}

// FuncDef 函数定义
type FuncDef struct {
	Public     bool
	ExternName string
	Name       string
	Params     []*Param
	Ret        Type
	Body       util.Option[*Block]
}

func (self FuncDef) GetPublic() bool {
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
	Public     bool
	Scope *StructDef
	Name       string
	SelfParam *Param
	Params     []*Param
	Ret        Type
	Body       *Block
}

func (self MethodDef) GetPublic() bool {
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
