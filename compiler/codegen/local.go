package codegen

import (
	"github.com/kkkunny/Sim/compiler/cir"
	"github.com/kkkunny/Sim/compiler/hir"
)

func (c *CodeGenerator) buildBlock(b *hir.Block) *cir.Block {
	stmts := make([]cir.Local, len(b.Stmts))
	for i, s := range b.Stmts {
		stmts[i] = c.buildLocal(s)
	}
	return &cir.Block{Stmts: stmts}
}

func (c *CodeGenerator) buildLocal(local hir.Local) cir.Local {
	switch local := local.(type) {
	case *hir.Block:
		return c.buildBlock(local)
	case *hir.Return:
		return c.buildReturn(local)
	case hir.Expr:
		return c.buildExpr(local)
	default:
		panic("unreachable")
	}
}

func (c *CodeGenerator) buildReturn(r *hir.Return) *cir.Return {
	var value cir.Expr
	if v, ok := r.Value.Value(); ok {
		value = c.buildExpr(v)
	}
	return &cir.Return{Value: value}
}
