package ast

import (
	"github.com/kkkunny/stl/container/linkedlist"

	"github.com/kkkunny/Sim/reader"
	"github.com/kkkunny/Sim/token"
	"github.com/kkkunny/Sim/util"
)

// Stmt 语句ast
type Stmt interface {
	Ast
	stmt()
}

// Block 代码块
type Block struct {
	Begin reader.Position
	Stmts linkedlist.LinkedList[Stmt]
	End   reader.Position
}

func (self *Block) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (*Block) stmt() {}

// Return 函数返回
type Return struct {
	Begin reader.Position
	Value util.Option[Expr]
}

func (self *Return) Position() reader.Position {
	v, ok := self.Value.Value()
	if !ok {
		return self.Begin
	}
	return reader.MixPosition(self.Begin, v.Position())
}

func (*Return) stmt() {}

// Variable 变量定义
type Variable struct {
	Begin reader.Position
	Name  token.Token
	Type  Type
	Value Expr
}

func (self *Variable) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.Value.Position())
}

func (*Variable) stmt() {}

// If if
type If struct {
	Begin reader.Position
	Cond  Expr
	Body  *Block
}

func (self *If) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.Body.Position())
}

func (*If) stmt() {}
