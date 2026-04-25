package codegen

import (
	"github.com/kkkunny/stl/container/optional"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/cir"
	"github.com/kkkunny/Sim/compiler/hir/stmts"
	"github.com/kkkunny/Sim/compiler/hir/types"
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
	if stlval.Is[types.FuncType](l.GetType()) {
		f := c.genNativeFunc(l.Value.(*stmts.Func))
		c.idents[l] = f.Decl
		if l.Name == "main" {
			f.Decl.Name = "sim_main"
		}
		return
	}

	t := c.genType(l.GetType())
	decl := cir.BuildStmt(c.builder, cir.NewVarDecl(t, ""))
	c.idents[l] = decl
	decl.Value = optional.Some(c.genExpr(l.Value))
}
