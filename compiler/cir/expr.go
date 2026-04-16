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
	Not UnaryOp `enum:"NOT"`
}]()

type UnaryExpr struct {
	Op   UnaryOp
	Expr Expr
}

func (*UnaryExpr) local() {}
func (*UnaryExpr) expr()  {}

type BinaryOp string

var BinaryOpEnum = enum.New[struct {
	Add BinaryOp `enum:"ADD"`
	Sub BinaryOp `enum:"SUB"`
	Mul BinaryOp `enum:"MUL"`
	Quo BinaryOp `enum:"QUO"`
	Rem BinaryOp `enum:"REM"`
	And BinaryOp `enum:"AND"`
	Or  BinaryOp `enum:"OR"`
	Xor BinaryOp `enum:"XOR"`
}]()

type BinaryExpr struct {
	Op    BinaryOp
	Left  Expr
	Right Expr
}

func (*BinaryExpr) local() {}
func (*BinaryExpr) expr()  {}
