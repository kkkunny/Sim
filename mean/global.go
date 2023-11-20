package mean

import (
	"github.com/kkkunny/stl/container/linkedhashmap"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/util"
)

// Global 全局
type Global interface {
	GetPublic() bool
}

// FuncDef 函数定义
type FuncDef struct {
	Public bool
	Name   string
	Params []*Param
	Ret    Type
	Body   util.Option[*Block]
}

func (self FuncDef) GetPublic() bool {
	return self.Public
}

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
	Public bool
	Name   string
	Fields linkedhashmap.LinkedHashMap[string, Type]
}

func (self StructDef) GetPublic() bool {
	return self.Public
}

func (self StructDef) String() string {
	return self.Name
}

func (self *StructDef) Equal(dst Type) bool {
	t, ok := dst.(*StructDef)
	if !ok {
		return false
	}
	return self == t
}

func (self *StructDef) AssignableTo(dst Type) bool {
	if self.Equal(dst) {
		return true
	}
	if ut, ok := dst.(*UnionType); ok {
		if ut.Elems.ContainKey(self.String()) {
			return true
		}
	}
	return false
}

// Variable 变量定义
type Variable struct {
	Public bool
	Mut    bool
	Type   Type
	Name   string
	Value  Expr
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
