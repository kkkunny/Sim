package codegen

import (
	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/cir"
)

func (c *CodeGenerator) buildFunc(fn *ast.FuncDecl) *cir.FuncDecl {
	params := make([]*cir.ParamDecl, len(fn.Params))
	for i, p := range fn.Params {
		params[i] = &cir.ParamDecl{
			Name: p.Name.OriginText,
			Type: c.buildType(p.Type),
		}
	}

	var returnType cir.Type = &cir.VoidType{}
	if rt, ok := fn.ReturnType.Value(); ok {
		returnType = c.buildType(rt)
	}

	var body *cir.Block
	if b, ok := fn.Body.Value(); ok {
		body = c.buildBlock(b)
	}

	return &cir.FuncDecl{
		Name:       fn.Name.OriginText,
		Params:     params,
		ReturnType: returnType,
		Body:       body,
	}
}
