package ast

import (
	"fmt"

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

func (e *IdentExpr) String() string {
	return e.Name.OriginText
}

type IntegerExpr struct {
	Value token.Token
}

func (e *IntegerExpr) local() {}

func (e *IntegerExpr) expr() {}

func (e *IntegerExpr) String() string {
	return e.Value.OriginText
}

type BinaryExpr struct {
	Op    token.Token
	Left  Expr
	Right Expr
}

func (e *BinaryExpr) local() {}

func (e *BinaryExpr) expr() {}

func (e *BinaryExpr) String() string {
	return fmt.Sprintf("%s %s %s", e.Left.String(), e.Op.OriginText, e.Right.String())
}
