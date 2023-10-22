package ast

import (
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
	Begin  reader.Position
	Name   token.Token
	Params []pair.Pair[token.Token, Type]
	Ret    util.Option[Type]
	Body   *Block
}

func (self *FuncDef) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.Body.Position())
}

func (self *FuncDef) global() {}

// StructDef 结构体定义
type StructDef struct {
	Begin  reader.Position
	Name   token.Token
	Fields []pair.Pair[token.Token, Type]
	End    reader.Position
}

func (self *StructDef) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (self *StructDef) global() {}
