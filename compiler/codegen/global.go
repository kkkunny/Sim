package codegen

import (
	"github.com/kkkunny/stl/container/optional"

	"github.com/kkkunny/Sim/compiler/cir"
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
	funcExpr := l.Value.(*hir.FuncExpr)

	returnType := c.buildType(funcExpr.ReturnType)

	params := make([]*cir.ParamDecl, len(funcExpr.Params))
	for i, p := range funcExpr.Params {
		params[i] = &cir.ParamDecl{
			Name: p.Name,
			Type: c.buildType(p.Type),
		}
	}

	f := c.builder.BuildFuncDecl(l.Name, returnType, params)

	if b, ok := funcExpr.Body.Value(); ok {
		f.Body = optional.Some(c.buildBlock(b))
	}
}
