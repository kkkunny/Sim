package ast

import (
	"github.com/kkkunny/stl/container/dynarray"
	"github.com/kkkunny/stl/container/pair"

	"github.com/kkkunny/Sim/reader"
	"github.com/kkkunny/Sim/token"
	"github.com/kkkunny/Sim/util"
)

// Global 全局ast
type Global interface {
	Ast
	global()
}

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

// StructDef 结构体定义
type StructDef struct {
	Begin  reader.Position
	Public bool
	Name   token.Token
	Fields []pair.Pair[token.Token, Type]
	End    reader.Position
}

func (self *StructDef) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (*StructDef) global() {}

// Variable 变量定义
type Variable struct {
	Attrs   []Attr
	Begin   reader.Position
	Public  bool
	Mutable bool
	Name    token.Token
	Type    Type
	Value   Expr
}

func (self *Variable) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.Value.Position())
}

func (*Variable) stmt() {}

func (*Variable) global() {}

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
