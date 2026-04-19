package cir

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
}

func (*IdentExpr) local() {}
func (*IdentExpr) expr()  {}

func (e *IdentExpr) print(p *printer) {
	p.WriteString(e.Name)
}

type IntegerExpr struct {
	Value *big.Int
}

func (*IntegerExpr) local() {}
func (*IntegerExpr) expr()  {}

func (e *IntegerExpr) print(p *printer) {
	p.WriteString(e.Value.String())
}

type FloatExpr struct {
	Value *big.Float
}

func (*FloatExpr) local() {}
func (*FloatExpr) expr()  {}

func (e *FloatExpr) print(p *printer) {
	p.WriteString(e.Value.String())
}

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

func (e *UnaryExpr) print(p *printer) {
	p.WriteString(string(e.Op))
	p.WriteString("(")
	p.WriteBy(e.Expr)
	p.WriteString(")")
}

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

func (e *BinaryExpr) print(p *printer) {
	p.WriteString(string(e.Op))
	p.WriteString("(")
	p.WriteBy(e.Left)
	p.WriteString(", ")
	p.WriteBy(e.Right)
	p.WriteString(")")
}

type FuncExpr struct {
	Decl *FuncDecl
}

func (*FuncExpr) local() {}
func (*FuncExpr) expr()  {}

func (e *FuncExpr) print(p *printer) {
	p.WriteString(e.Decl.Name)
}

type MacroExpr struct {
	Name string
	Args []Expr
}

func NewMacroExpr(name string, args ...Expr) *MacroExpr {
	return &MacroExpr{
		Name: name,
		Args: args,
	}
}

func (*MacroExpr) local() {}
func (*MacroExpr) expr()  {}

func (e *MacroExpr) print(p *printer) {
	p.WriteString(e.Name)
	p.WriteString("(")
	for i, arg := range e.Args {
		p.WriteBy(arg)
		if i < len(e.Args)-1 {
			p.WriteString(", ")
		}
	}
	p.WriteString(")")
}
