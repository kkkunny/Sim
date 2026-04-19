package codegen

import (
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/hir"
)

func (c *CodeGenerator) buildGlobal(global hir.Global) {
	switch global := global.(type) {
	case *hir.Let:
		c.buildGlobalLet(global)
	default:
		panic("unreachable")
	}
}

func (c *CodeGenerator) buildGlobalLet(l *hir.Let) {
	if stlval.Is[*hir.FuncType](l.GetType()) {
		f := c.buildNativeFunc(l.Value.(*hir.Func))
		c.idents[l] = f.Decl
		if l.Name == "main" {
			f.Decl.Name = "sim_main"
		}
		return
	}

	t := c.buildType(l.GetType())
	value := c.buildExpr(l.Value)
	c.builder.BuildVarDecl(t, "", value)
}
