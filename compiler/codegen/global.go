package codegen

import (
	"github.com/kkkunny/Sim/compiler/cir"
	"github.com/kkkunny/Sim/compiler/hir"
)

func (c *CodeGenerator) buildGlobal(global hir.Global) cir.Global {
	switch global := global.(type) {
	case *hir.Let:
		return c.buildGlobalLet(global)
	default:
		panic("unreachable")
	}
}

func (c *CodeGenerator) buildGlobalLet(l *hir.Let) *cir.FuncDecl {
	funcExpr := l.Value.(*hir.FuncExpr)

	returnType := c.buildType(funcExpr.ReturnType)

	params := make([]*cir.ParamDecl, len(funcExpr.Params))
	for i, p := range funcExpr.Params {
		params[i] = &cir.ParamDecl{
			Name: p.Name,
			Type: c.buildType(p.Type),
		}
	}

	var body *cir.Block
	if b, ok := funcExpr.Body.Value(); ok {
		body = c.buildBlock(b)
	}

	return &cir.FuncDecl{
		Name:       l.Name,
		Params:     params,
		ReturnType: returnType,
		Body:       body,
	}
}
