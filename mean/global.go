package mean

import (
	"github.com/kkkunny/stl/container/linkedhashmap"
	"github.com/samber/lo"
)

// Global 全局
type Global interface {
	global()
}

// FuncDef 函数定义
type FuncDef struct {
	Name   string
	Params []*Param
	Ret    Type
	Body   *Block
}

func (*FuncDef) global() {}

func (*FuncDef) stmt() {}

func (self *FuncDef) GetType() Type {
	params := lo.Map(self.Params, func(item *Param, index int) Type {
		return item.GetType()
	})
	return &FuncType{
		Ret:    self.Ret,
		Params: params,
	}
}

func (self *FuncDef) Mutable() bool {
	return false
}

func (*FuncDef) ident() {}

// StructDef 结构体定义
type StructDef struct {
	Name   string
	Fields linkedhashmap.LinkedHashMap[string, Type]
}

func (*StructDef) global() {}

func (self *StructDef) Equal(dst Type) bool {
	t, ok := dst.(*StructDef)
	if !ok {
		return false
	}
	return self == t
}
