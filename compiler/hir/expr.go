package hir

import (
	"math/big"

	"github.com/kkkunny/stl/enum"
)

type Expr interface {
	Local
	expr()
}

type IdentExpr struct {
	Name string
	Type Type
}

func (*IdentExpr) expr()  {}
func (*IdentExpr) local() {}

func (e *IdentExpr) print(p *printer) {
	p.WriteString(e.Name)
}

type IntegerExpr struct {
	Value *big.Int
	Type  Type
}

func (*IntegerExpr) expr()  {}
func (*IntegerExpr) local() {}

func (e *IntegerExpr) print(p *printer) {
	p.WriteString(e.Value.String())
}

type UnaryOp string

var UnaryOpEnum = enum.New[struct {
	Not UnaryOp `enum:"!"`
}]()

type UnaryExpr struct {
	Op   UnaryOp
	Expr Expr
}

func (*UnaryExpr) expr()  {}
func (*UnaryExpr) local() {}

func (e *UnaryExpr) print(p *printer) {
	p.WriteString("(")
	p.WriteString(string(e.Op))
	p.WriteBy(e.Expr)
	p.WriteString(")")
}

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

func (*BinaryExpr) expr()  {}
func (*BinaryExpr) local() {}

func (e *BinaryExpr) print(p *printer) {
	p.WriteString("(")
	p.WriteBy(e.Left)
	p.WriteString(" ")
	p.WriteString(string(e.Op))
	p.WriteString(" ")
	p.WriteBy(e.Right)
	p.WriteString(")")
}
