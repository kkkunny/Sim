package codegen

import (
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
	f := c.buildNativeFunc(l.Value.(*hir.Func))
	c.idents[l] = f.Decl
	if l.Name == "main" {
		f.Decl.Name = "sim_main"
	}
}
