package ast

import (
	"github.com/kkkunny/stl/container/dynarray"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/reader"
	"github.com/kkkunny/Sim/token"
	"github.com/kkkunny/Sim/util"
)

// Global 全局ast
type Global interface {
	Ast
	global()
}

// StructDef 结构体定义
type StructDef struct {
	Begin  reader.Position
	Public bool
	Name   token.Token
	Fields []lo.Tuple3[bool, token.Token, Type]
	End    reader.Position
}

func (self *StructDef) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (*StructDef) global() {}


type VariableDef interface {
	Global
	Stmt
	variable()
}

type VarDef struct {
	Mutable bool
	Name    token.Token
	Type    util.Option[Type]
}

// SingleVariableDef 单变量定义
type SingleVariableDef struct {
	Attrs   []Attr
	Begin   reader.Position
	Public bool
	Var    VarDef
	Value  util.Option[Expr]
}

func (self *SingleVariableDef) Position() reader.Position {
	if v, ok := self.Value.Value(); ok{
		return reader.MixPosition(self.Begin, v.Position())
	}else{
		return reader.MixPosition(self.Begin, self.Var.Type.MustValue().Position())
	}
}

func (*SingleVariableDef) stmt() {}

func (*SingleVariableDef) global() {}

func (*SingleVariableDef) variable() {}

// MultipleVariableDef 多变量定义
type MultipleVariableDef struct {
	Attrs   []Attr
	Begin   reader.Position
	Public  bool
	Vars []VarDef
	Value   util.Option[Expr]
	End reader.Position
}

func (self *MultipleVariableDef) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (*MultipleVariableDef) stmt() {}

func (*MultipleVariableDef) global() {}

func (*MultipleVariableDef) variable() {}

func (self *MultipleVariableDef) ToSingleList()[]*SingleVariableDef{
	return lo.Map(self.Vars, func(item VarDef, _ int) *SingleVariableDef {
		return &SingleVariableDef{
			Attrs: self.Attrs,
			Begin: self.Begin,
			Public: self.Public,
			Var: item,
			Value: util.None[Expr](),
		}
	})
}

// Import 包导入
type Import struct {
	Begin reader.Position
	Paths dynarray.DynArray[token.Token]
	Alias util.Option[token.Token]
}

func (self *Import) Position() reader.Position {
	if alias, ok := self.Alias.Value(); ok {
		return reader.MixPosition(self.Begin, alias.Position)
	}
	return reader.MixPosition(self.Begin, self.Paths.Back().Position)
}

func (*Import) global() {}

// TypeAlias 类型别名
type TypeAlias struct {
	Begin  reader.Position
	Public bool
	Name   token.Token
	Type   Type
}

func (self *TypeAlias) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.Type.Position())
}

func (*TypeAlias) global() {}

// FuncDef 函数定义
type FuncDef struct {
	Attrs    []Attr
	Begin    reader.Position
	Public   bool
	Name     token.Token
	Params   []Param
	ParamEnd reader.Position
	Ret      util.Option[Type]
	Body     util.Option[*Block]
}

func (self *FuncDef) Position() reader.Position {
	if b, ok := self.Body.Value(); ok {
		return reader.MixPosition(self.Begin, b.Position())
	} else if r, ok := self.Ret.Value(); ok {
		return reader.MixPosition(self.Begin, r.Position())
	} else {
		return reader.MixPosition(self.Begin, self.ParamEnd)
	}
}

func (*FuncDef) global() {}

// MethodDef 方法定义
type MethodDef struct {
	Attrs    []Attr
	Begin    reader.Position
	Public   bool
	ScopeMutable bool
	Scope token.Token
	Name     token.Token
	Params   []Param
	Ret      util.Option[Type]
	Body     *Block
}

func (self *MethodDef) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.Body.Position())
}

func (*MethodDef) global() {}

// GenericFuncDef 泛型函数定义
type GenericFuncDef struct {
	Attrs    []Attr
	Begin    reader.Position
	Public   bool
	Name     GenericNameDef
	Params   []Param
	Ret      util.Option[Type]
	Body     *Block
}

func (self *GenericFuncDef) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.Body.Position())
}

func (*GenericFuncDef) global() {}

// GenericStructDef 泛型结构体定义
type GenericStructDef struct {
	Begin  reader.Position
	Public bool
	Name   GenericNameDef
	Fields []lo.Tuple3[bool, token.Token, Type]
	End    reader.Position
}

func (self *GenericStructDef) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (*GenericStructDef) global() {}

// GenericStructMethodDef 泛型结构体方法定义
type GenericStructMethodDef struct {
	Attrs    []Attr
	Begin    reader.Position
	Public   bool
	ScopeMutable bool
	Scope GenericNameDef
	Name     token.Token
	Params   []Param
	Ret      util.Option[Type]
	Body     *Block
}

func (self *GenericStructMethodDef) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.Body.Position())
}

func (*GenericStructMethodDef) global() {}

// GenericMethodDef 泛型方法定义
type GenericMethodDef struct {
	Attrs    []Attr
	Begin    reader.Position
	Public   bool
	ScopeMutable bool
	Scope token.Token
	Name     GenericNameDef
	Params   []Param
	Ret      util.Option[Type]
	Body     *Block
}

func (self *GenericMethodDef) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.Body.Position())
}

func (*GenericMethodDef) global() {}

// GenericStructGenericMethodDef 泛型结构体泛型方法定义
type GenericStructGenericMethodDef struct {
	Attrs    []Attr
	Begin    reader.Position
	Public   bool
	ScopeMutable bool
	Scope GenericNameDef
	Name     GenericNameDef
	Params   []Param
	Ret      util.Option[Type]
	Body     *Block
}

func (self *GenericStructGenericMethodDef) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.Body.Position())
}

func (*GenericStructGenericMethodDef) global() {}
