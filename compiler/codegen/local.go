package codegen

import (
	"math/big"

	"github.com/kkkunny/stl/container/either"
	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/cir"
	"github.com/kkkunny/Sim/compiler/hir/stmts"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

func (c *CodeGenerator) genLocal(local stmts.Local) {
	switch local := local.(type) {
	case *stmts.Block:
		c.genBlock(false, local, nil)
	case *stmts.Return:
		c.genReturn(local)
	case *stmts.Let:
		c.genLocalLet(local)
	case stmts.Expr:
		cir.BuildStmt(c.builder, cir.NewExpr(c.genExpr(local)))
	case *stmts.If:
		c.genIf(local, true)
	case *stmts.While:
		c.genWhile(local)
	case *stmts.For:
		c.genFor(local)
	default:
		panic("unreachable")
	}
}

func (c *CodeGenerator) buildFuncBlock(f func()) *cir.Block {
	savedVarCount := c.builder.SaveFuncVarCount()
	defer c.builder.RestoreFuncVarCount(savedVarCount)

	block := c.buildBlock(true, f)
	return block
}

func (c *CodeGenerator) buildBlock(flat bool, f func()) *cir.Block {
	prevBlock, _ := c.builder.CurrentAt()

	block := stlval.IfLazy(flat, func() *cir.Block {
		return cir.NewBlock()
	}, func() *cir.Block {
		return cir.BuildStmt(c.builder, cir.NewBlock())
	})
	c.builder.MoveTo(block)
	defer c.builder.MoveTo(prevBlock)

	if f != nil {
		f()
	}
	return block
}

func (c *CodeGenerator) genFuncBlock(b *stmts.Block, initFn func()) *cir.Block {
	return c.buildFuncBlock(func() {
		if initFn != nil {
			initFn()
		}
		for _, s := range b.Stmts {
			c.genLocal(s)
		}
	})
}

func (c *CodeGenerator) genBlock(flat bool, b *stmts.Block, initFn func()) *cir.Block {
	return c.buildBlock(flat, func() {
		if initFn != nil {
			initFn()
		}
		for _, s := range b.Stmts {
			c.genLocal(s)
		}
	})
}

func (c *CodeGenerator) genReturn(local *stmts.Return) *cir.Return {
	if v, ok := local.Value.Value(); ok {
		return cir.BuildStmt(c.builder, cir.NewReturn(c.genExpr(v)))
	} else {
		return cir.BuildStmt(c.builder, cir.NewReturn())
	}
}

func (c *CodeGenerator) genLocalLet(local *stmts.Let) *cir.VarDecl {
	typ := c.genType(local.Value.GetType())
	value := c.genExpr(local.Value)
	v := cir.BuildStmt(c.builder, cir.NewVarDecl(typ, "", value))
	c.idents[local] = v
	return v
}

func (c *CodeGenerator) genIf(l *stmts.If, isRoot bool) *cir.If {
	cond := c.genExpr(l.Condition)
	body := c.genBlock(true, l.Body, nil)
	var next []either.Either[*cir.If, *cir.Block]
	if nextIf, ok := l.Else.Value(); ok {
		if elseif, ok := nextIf.TryLeft(); ok {
			next = append(next, either.Left[*cir.If, *cir.Block](c.genIf(elseif, false)))
		} else {
			next = append(next, either.Right[*cir.If, *cir.Block](c.genBlock(true, nextIf.Right(), nil)))
		}
	}

	if !isRoot {
		return &cir.If{
			Condition: cond,
			Body:      body,
			Else:      optional.UnEmpty(stlslices.Last(next)),
		}
	}
	return cir.BuildStmt(c.builder, cir.NewIf(cond, body, next...))
}

func (c *CodeGenerator) genWhile(l *stmts.While) *cir.While {
	cond := c.genExpr(l.Condition)
	body := c.genBlock(true, l.Body, nil)
	return cir.BuildStmt(c.builder, cir.NewWhile(cond, body))
}

func (c *CodeGenerator) genFor(l *stmts.For) *cir.For {
	init := cir.BuildStmt(c.builder, cir.NewVarDecl(cir.I64, "", cir.NewInteger(big.NewInt(0))))
	rangv := c.genExpr(l.Range)
	if l.Range.Temporary() {
		rangvar := cir.BuildStmt(c.builder, cir.NewVarDecl(c.genType(l.Range.GetType()), "", rangv))
		rangv = cir.NewIdentExpr(rangvar.Name)
	}
	at := l.Range.GetType().(types.ArrayType)
	cond := cir.NewBinary(cir.BinaryOpEnum.Lt, cir.NewIdentExpr(init.Name), cir.NewInteger(at.GetSize()))
	action := cir.NewUnary(cir.UnaryOpEnum.SelfAdd, cir.NewIdentExpr(init.Name))
	body := c.genBlock(true, l.Body, func() {
		v := cir.BuildStmt(c.builder, cir.NewVarDecl(c.genType(at), "", cir.NewOffset(cir.NewGetMember(rangv, "array"), cir.NewIdentExpr(init.Name))))
		c.idents[l.Var] = v
	})
	return cir.BuildStmt(c.builder, cir.NewFor(optional.None[*cir.VarDecl](), optional.Some[cir.Expr](cond), optional.Some[cir.Expr](action), body))
}
