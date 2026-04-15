package codegen

import (
	"github.com/kkkunny/stl/container/optional"

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
	case *hir.Let:
		return c.buildLocalLet(local)
	default:
		panic("unreachable")
	}
}

func (c *CodeGenerator) buildReturn(r *hir.Return) *cir.Return {
	var value optional.Optional[cir.Expr]
	if v, ok := r.Value.Value(); ok {
		value = optional.Some(c.buildExpr(v))
	}
	return &cir.Return{Value: value}
}

func (c *CodeGenerator) buildLocalLet(l *hir.Let) *cir.VarDecl {
	typ := c.buildType(l.Value.GetType())
	value := c.buildExpr(l.Value)
	return &cir.VarDecl{
		Type:  typ,
		Name:  l.Name,
		Value: optional.Some(value),
	}
}
