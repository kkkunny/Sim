package codegen

import (
	"github.com/kkkunny/stl/container/optional"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/hir/stmts"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

func (c *CodeGenerator) buildGlobal(global stmts.Global) {
	switch global := global.(type) {
	case *stmts.Let:
		c.buildGlobalLet(global)
	default:
		panic("unreachable")
	}
}

func (c *CodeGenerator) buildGlobalLet(l *stmts.Let) {
	if stlval.Is[types.FuncType](l.GetType()) {
		f := c.buildNativeFunc(l.Value.(*stmts.Func))
		c.idents[l] = f.Decl
		if l.Name == "main" {
			f.Decl.Name = "sim_main"
		}
		return
	}

	t := c.buildType(l.GetType())
	decl := c.builder.BuildVarDecl(t, "")
	c.idents[l] = decl
	decl.Value = optional.Some(c.buildExpr(l.Value))
}
