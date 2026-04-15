package codegen

import (
	"github.com/kkkunny/Sim/compiler/cir"
	"github.com/kkkunny/Sim/compiler/hir"
)

func (c *CodeGenerator) buildFunc(fn *hir.FuncDecl) *cir.FuncDecl {
	params := make([]*cir.ParamDecl, len(fn.Params))
	for i, p := range fn.Params {
		params[i] = &cir.ParamDecl{
			Name: p.Name,
			Type: c.buildType(p.Type),
		}
	}

	var returnType cir.Type = &cir.VoidType{}
	if fn.ReturnType != nil {
		returnType = c.buildType(fn.ReturnType)
	}

	var body *cir.Block
	if b, ok := fn.Body.Value(); ok {
		body = c.buildBlock(b)
	}

	return &cir.FuncDecl{
		Name:       fn.Name,
		Params:     params,
		ReturnType: returnType,
		Body:       body,
	}
}
