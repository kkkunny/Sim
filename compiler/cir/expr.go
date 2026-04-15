package cir

import "github.com/kkkunny/stl/enum"

type Expr interface {
	Local
	expr()
}

type IdentExpr struct {
	Name string
}

func (*IdentExpr) local() {}
func (*IdentExpr) expr()  {}

type IntegerExpr struct {
	Value string
}

func (*IntegerExpr) local() {}
func (*IntegerExpr) expr()  {}

type UnaryOp string

var UnaryOpEnum = enum.New[struct {
	Not UnaryOp `enum:"!"`
}]()

type UnaryExpr struct {
	Op   UnaryOp
	Expr Expr
}

func (*UnaryExpr) local() {}
func (*UnaryExpr) expr()  {}

type BinaryOp string

var BinaryOpEnum = enum.New[struct {
	Add BinaryOp `enum:"+"`
	Sub BinaryOp `enum:"-"`
	Mul BinaryOp `enum:"*"`
	Quo BinaryOp `enum:"/"`
	Rem BinaryOp `enum:"%"`
	And BinaryOp `enum:"&"`
	Or  BinaryOp `enum:"|"`
	Xor BinaryOp `enum:"^"`
}]()

type BinaryExpr struct {
	Op    BinaryOp
	Left  Expr
	Right Expr
}

func (*BinaryExpr) local() {}
func (*BinaryExpr) expr()  {}
