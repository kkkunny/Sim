package cir

import "github.com/kkkunny/stl/container/optional"

type Local interface {
	local()
}

type Block struct {
	Stmts []Local
}

func (*Block) local() {}

type Return struct {
	Value optional.Optional[Expr]
}

func (*Return) local() {}

type VarDecl struct {
	Type  Type
	Name  string
	Value optional.Optional[Expr]
}

func (*VarDecl) local() {}
