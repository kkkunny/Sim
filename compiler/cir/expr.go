package cir

type Expr interface {
	Local
	expr()
}

type IdentExpr struct {
	Name string
}

func (*IdentExpr) local() {}
func (*IdentExpr) expr()  {}
