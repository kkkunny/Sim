package codegen

import (
	"github.com/kkkunny/stl/container/optional"

	"github.com/kkkunny/Sim/compiler/cir"
	"github.com/kkkunny/Sim/compiler/hir/stmts"
)

func (c *CodeGenerator) genGlobal(global stmts.Global) {
	switch global := global.(type) {
	case *stmts.Let:
		c.genGlobalLet(global)
	default:
		panic("unreachable")
	}
}

func (c *CodeGenerator) genGlobalLet(l *stmts.Let) {
	if expr, ok := l.Value.(*stmts.Func); ok {
		decl := c.genNativeFuncDecl(expr)
		c.idents[l] = decl
		if b, ok := expr.Body.Value(); ok {
			prevFunc := c.currentFunc
			c.currentFunc = expr
			decl.Body = optional.Some(c.genFuncBlock(b, nil))
			c.currentFunc = prevFunc
		}
		if l.Name == "main" {
			decl.Name = "sim_main"
		}
		return
	}

	t := c.genType(l.GetType())
	decl := cir.BuildStmt(c.builder, cir.NewVarDecl(t, ""))
	c.idents[l] = decl
	decl.Value = optional.Some(c.genExpr(l.Value))
}
