package cir

import "github.com/kkkunny/stl/enum"

type BinaryOp string

var BinaryOpEnum = enum.New[struct {
	Add BinaryOp `enum:"+"`
	Sub BinaryOp `enum:"-"`
	Mul BinaryOp `enum:"*"`
	Quo BinaryOp `enum:"/"`
}]()

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

type BinaryExpr struct {
	Op    BinaryOp
	Left  Expr
	Right Expr
}

func (*BinaryExpr) local() {}
func (*BinaryExpr) expr()  {}
