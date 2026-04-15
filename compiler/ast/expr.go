package ast

import (
	"github.com/kkkunny/stl/container/optional"

	"github.com/kkkunny/Sim/compiler/token"
)

type Expr interface {
	Local
	expr()
}

type IdentExpr struct {
	Name token.Token
}

func (e *IdentExpr) local() {}
func (e *IdentExpr) expr()  {}

func (e *IdentExpr) print(p *printer) {
	p.WriteToken(e.Name)
}

type IntegerExpr struct {
	Value token.Token
}

func (e *IntegerExpr) local() {}
func (e *IntegerExpr) expr()  {}

func (e *IntegerExpr) print(p *printer) {
	p.WriteToken(e.Value)
}

type UnaryExpr struct {
	Op   token.Token
	Expr Expr
}

func (e *UnaryExpr) local() {}
func (e *UnaryExpr) expr()  {}

func (e *UnaryExpr) print(p *printer) {
	p.WriteString("(")
	p.WriteToken(e.Op)
	p.WriteBy(e.Expr)
	p.WriteString(")")
}

type BinaryExpr struct {
	Op    token.Token
	Left  Expr
	Right Expr
}

func (e *BinaryExpr) local() {}
func (e *BinaryExpr) expr()  {}

func (e *BinaryExpr) print(p *printer) {
	p.WriteString("(")
	p.WriteBy(e.Left)
	p.WriteString(" ")
	p.WriteToken(e.Op)
	p.WriteString(" ")
	p.WriteBy(e.Right)
	p.WriteString(")")
}

type FuncExpr struct {
	Params     []*ParamDecl
	ReturnType optional.Optional[Type]
	Body       optional.Optional[*Block]
}

func (e *FuncExpr) local() {}
func (e *FuncExpr) expr()  {}

func (e *FuncExpr) print(p *printer) {
	p.WriteString("(")
	for i, param := range e.Params {
		if i > 0 {
			p.WriteString(", ")
		}
		p.WriteBy(param)
	}
	p.WriteString(")")
	if rt, ok := e.ReturnType.Value(); ok {
		p.WriteString(" -> ")
		p.WriteBy(rt)
	}
	if body, ok := e.Body.Value(); ok {
		p.WriteString(" ")
		p.WriteBy(body)
	}
}
