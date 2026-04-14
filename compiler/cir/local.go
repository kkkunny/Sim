package cir

type Local interface {
	local()
}

type Block struct {
	Stmts []Local
}

func (*Block) local() {}

type Return struct {
	Value Expr
}

func (*Return) local() {}
