package hir

import (
	"math/big"

	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
	"github.com/kkkunny/stl/enum"
)

type Expr interface {
	Local
	expr()
	GetType() Type
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

func (e *IdentExpr) GetType() Type {
	return e.Type
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

func (e *IntegerExpr) GetType() Type {
	return I32
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

func (e *UnaryExpr) GetType() Type {
	return e.Expr.GetType()
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

func (e *BinaryExpr) GetType() Type {
	return e.Left.GetType()
}

type FuncExpr struct {
	Params     []*ParamDecl
	ReturnType Type
	Body       optional.Optional[*Block]
}

func (*FuncExpr) expr()  {}
func (*FuncExpr) local() {}

func (e *FuncExpr) print(p *printer) {
	p.WriteString("(")
	for i, param := range e.Params {
		if i > 0 {
			p.WriteString(", ")
		}
		p.WriteBy(param)
	}
	p.WriteString(")")
	p.WriteString(" -> ")
	p.WriteBy(e.ReturnType)
	if body, ok := e.Body.Value(); ok {
		p.WriteString(" ")
		p.WriteBy(body)
	}
}

func (e *FuncExpr) GetType() Type {
	params := stlslices.Map(e.Params, func(_ int, param *ParamDecl) Type {
		return param.Type
	})
	return NewFuncType(e.ReturnType, params...)
}
