package ast

import "github.com/kkkunny/Sim/compiler/token"

type Expr interface {
	Local
	expr()
}

type IdentExpr struct {
	Name token.Token
}

func (e *IdentExpr) local() {}

func (e *IdentExpr) expr() {}

func (e *IdentExpr) String() string {
	return e.Name.OriginText
}
