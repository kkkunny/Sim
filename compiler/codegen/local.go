package codegen

import (
	"github.com/kkkunny/stl/container/either"
	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/cir"
	"github.com/kkkunny/Sim/compiler/hir/stmts"
)

func (c *CodeGenerator) buildLocal(local stmts.Local) {
	switch local := local.(type) {
	case *stmts.Block:
		c.buildBlock(local, nil)
	case *stmts.Return:
		c.buildReturn(local)
	case *stmts.Let:
		c.buildLocalLet(local)
	case stmts.Expr:
		c.builder.BuildExpr(c.buildExpr(local))
	case *stmts.If:
		c.buildIf(local, true)
	default:
		panic("unreachable")
	}
}

func (c *CodeGenerator) buildBlock(b *stmts.Block, initFn func()) *cir.Block {
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

func (c *CodeGenerator) buildFlatBlock(b *stmts.Block, initFn func()) *cir.Block {
	prevBlock, _ := c.builder.CurrentAt()
	block := &cir.Block{}
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

func (c *CodeGenerator) buildReturn(local *stmts.Return) *cir.Return {
	if v, ok := local.Value.Value(); ok {
		return c.builder.BuildReturn(c.buildExpr(v))
	} else {
		return c.builder.BuildReturn()
	}
}

func (c *CodeGenerator) buildLocalLet(local *stmts.Let) *cir.VarDecl {
	typ := c.buildType(local.Value.GetType())
	value := c.buildExpr(local.Value)
	v := c.builder.BuildVarDecl(typ, "", value)
	c.idents[local] = v
	return v
}

func (c *CodeGenerator) buildIf(l *stmts.If, isRoot bool) *cir.If {
	cond := c.buildExpr(l.Condition)
	body := c.buildFlatBlock(l.Body, nil)
	var next []either.Either[*cir.If, *cir.Block]
	if nextIf, ok := l.Else.Value(); ok {
		if elseif, ok := nextIf.TryLeft(); ok {
			next = append(next, either.Left[*cir.If, *cir.Block](c.buildIf(elseif, false)))
		} else {
			next = append(next, either.Right[*cir.If, *cir.Block](c.buildFlatBlock(nextIf.Right(), nil)))
		}
	}

	if !isRoot {
		return &cir.If{
			Condition: cond,
			Body:      body,
			Else:      optional.UnEmpty(stlslices.Last(next)),
		}
	}
	return c.builder.BuildIf(cond, body, next...)
}
