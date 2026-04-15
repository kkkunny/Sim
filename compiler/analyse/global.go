package analyse

import (
	"github.com/kkkunny/stl/container/optional"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/hir"
)

func (a *Analyzer) analyzeFunc(fn *ast.FuncDecl) *hir.FuncDecl {
	a.scope = &Scope{parent: a.scope, types: make(map[string]hir.Type)}

	var params []*hir.ParamDecl
	for _, p := range fn.Params {
		paramType := a.analyseType(p.Type)
		params = append(params, &hir.ParamDecl{
			Name: p.Name.OriginText,
			Type: paramType,
		})
		a.scope.types[p.Name.OriginText] = paramType
	}

	var returnType hir.Type
	if rt, ok := fn.ReturnType.Value(); ok {
		returnType = a.analyseType(rt)
	} else {
		returnType = hir.Unit
	}

	var body optional.Optional[*hir.Block]
	if b, ok := fn.Body.Value(); ok {
		body = optional.Some(a.analyzeBlock(b))
	}

	return &hir.FuncDecl{
		Name:       fn.Name.OriginText,
		Params:     params,
		ReturnType: returnType,
		Body:       body,
	}
}
