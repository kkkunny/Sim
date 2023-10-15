package ast

import (
	"github.com/kkkunny/Sim/reader"
	"github.com/kkkunny/Sim/token"
)

// Expr 表达式
type Expr interface {
	Stmt
	expr()
}

// Integer 整数
type Integer struct {
	Value token.Token
}

func (self *Integer) Position() reader.Position {
	return self.Value.Position
}

func (self *Integer) stmt() {}

func (self *Integer) expr() {}

// Float 浮点数
type Float struct {
	Value token.Token
}

func (self *Float) Position() reader.Position {
	return self.Value.Position
}

func (self *Float) stmt() {}

func (self *Float) expr() {}

// Binary 二元运算
type Binary struct {
	Left  Expr
	Opera token.Token
	Right Expr
}

func (self *Binary) Position() reader.Position {
	return reader.MixPosition(self.Left.Position(), self.Right.Position())
}

func (self *Binary) stmt() {}

func (self *Binary) expr() {}
