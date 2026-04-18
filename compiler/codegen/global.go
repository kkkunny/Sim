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
	c.buildFuncExpr(l.Value.(*hir.FuncExpr))
}
