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

// Unary 一元运算
type Unary struct {
	Opera token.Token
	Value Expr
}

func (self *Unary) Position() reader.Position {
	return reader.MixPosition(self.Opera.Position, self.Value.Position())
}

func (self *Unary) stmt() {}

func (self *Unary) expr() {}

// Boolean 布尔值
type Boolean struct {
	Value token.Token
}

func (self *Boolean) Position() reader.Position {
	return self.Value.Position
}

func (self *Boolean) stmt() {}

func (self *Boolean) expr() {}

// Ident 标识符
type Ident struct {
	Name token.Token
}

func (self *Ident) Position() reader.Position {
	return self.Name.Position
}

func (self *Ident) stmt() {}

func (self *Ident) expr() {}

// Call 调用
type Call struct {
	Func Expr
	End  reader.Position
}

func (self *Call) Position() reader.Position {
	return reader.MixPosition(self.Func.Position(), self.End)
}

func (self *Call) stmt() {}

func (self *Call) expr() {}

// Unit 单元
type Unit struct {
	Begin reader.Position
	Value Expr
	End   reader.Position
}

func (self *Unit) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (self *Unit) stmt() {}

func (self *Unit) expr() {}

// Covert 类型转换
type Covert struct {
	Value Expr
	Type  Type
}

func (self *Covert) Position() reader.Position {
	return reader.MixPosition(self.Value.Position(), self.Type.Position())
}

func (self *Covert) stmt() {}

func (self *Covert) expr() {}

// Array 数组
type Array struct {
	Type  *ArrayType
	Elems []Expr
	End   reader.Position
}

func (self *Array) Position() reader.Position {
	return reader.MixPosition(self.Type.Position(), self.End)
}

func (self *Array) stmt() {}

func (self *Array) expr() {}
