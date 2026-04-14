package codegen

import (
	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/cir"
)

func (c *CodeGenerator) buildLocal(local ast.Local) cir.Local {
	switch local := local.(type) {
	case *ast.Block:
		return c.buildBlock(local)
	case *ast.Return:
		return c.buildReturn(local)
	case ast.Expr:
		return c.buildExpr(local)
	default:
		panic("unreachable")
	}
}

func (c *CodeGenerator) buildBlock(b *ast.Block) *cir.Block {
	stmts := make([]cir.Local, len(b.Stmts))
	for i, s := range b.Stmts {
		stmts[i] = c.buildLocal(s)
	}
	return &cir.Block{Stmts: stmts}
}

func (c *CodeGenerator) buildReturn(r *ast.Return) *cir.Return {
	var value cir.Expr
	if v, ok := r.Value.Value(); ok {
		value = c.buildExpr(v)
	}
	return &cir.Return{Value: value}
}
