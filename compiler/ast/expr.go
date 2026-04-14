package ast

import (
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

func (e *IdentExpr) expr() {}

func (e *IdentExpr) print(p *printer) {
	p.WriteToken(e.Name)
}

type IntegerExpr struct {
	Value token.Token
}

func (e *IntegerExpr) local() {}

func (e *IntegerExpr) expr() {}

func (e *IntegerExpr) print(p *printer) {
	p.WriteToken(e.Value)
}

type BinaryExpr struct {
	Op    token.Token
	Left  Expr
	Right Expr
}

func (e *BinaryExpr) local() {}

func (e *BinaryExpr) expr() {}

func (e *BinaryExpr) print(p *printer) {
	p.WriteBy(e.Left)
	p.WriteString(" ")
	p.WriteToken(e.Op)
	p.WriteString(" ")
	p.WriteBy(e.Right)
}
