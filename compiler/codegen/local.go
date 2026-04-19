package codegen

import (
	"github.com/kkkunny/Sim/compiler/cir"
	"github.com/kkkunny/Sim/compiler/hir"
)

func (c *CodeGenerator) buildLocal(local hir.Local) {
	switch local := local.(type) {
	case *hir.Block:
		c.buildBlock(local, nil)
	case *hir.Return:
		c.buildReturn(local)
	case *hir.Let:
		c.buildLocalLet(local)
	case hir.Expr:
		c.buildExpr(local)
	default:
		panic("unreachable")
	}
}

func (c *CodeGenerator) buildBlock(b *hir.Block, initFn func()) *cir.Block {
	prevBlock, _ := c.builder.CurrentAt()
	block := c.builder.BuildBlock()
	c.builder.MoveTo(block)
	if initFn != nil {
		initFn()
	}
	for _, s := range b.Stmts {
		c.buildLocal(s)
	}
	c.builder.MoveTo(prevBlock)
	return block
}

func (c *CodeGenerator) buildReturn(r *hir.Return) *cir.Return {
	if v, ok := r.Value.Value(); ok {
		return c.builder.BuildReturn(c.buildExpr(v))
	} else {
		return c.builder.BuildReturn()
	}
}

func (c *CodeGenerator) buildLocalLet(l *hir.Let) *cir.VarDecl {
	typ := c.buildType(l.Value.GetType())
	value := c.buildExpr(l.Value)
	v := c.builder.BuildVarDecl(typ, "", value)
	c.idents[l] = v
	return v
}
