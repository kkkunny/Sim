package analyse

import (
	"github.com/kkkunny/stl/container/set"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/ast"
	errors "github.com/kkkunny/Sim/compiler/error"
	"github.com/kkkunny/Sim/compiler/hir/global"
	"github.com/kkkunny/Sim/compiler/hir/local"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

func (self *Analyser) analyseParam(node ast.Param, analysers ...typeAnalyser) *local.Param {
	return local.NewParam(
		!node.Mutable.IsNone(),
		stlval.TernaryAction(node.Name.IsNone(), func() string {
			return "_"
		}, func() string {
			return node.Name.MustValue().Source()
		}),
		self.analyseType(node.Type, analysers...),
	)
}

func (self *Analyser) analyseFuncDecl(node ast.FuncDecl, analysers ...typeAnalyser) *global.FuncDecl {
	name := node.Name.Source()

	paramNameSet := set.StdHashSetWith[string]()
	params := stlslices.Map(node.Params, func(_ int, e ast.Param) *local.Param {
		param := self.analyseParam(e, analysers...)
		if paramName, ok := param.GetName(); ok && !paramNameSet.Add(paramName) {
			errors.ThrowIdentifierDuplicationError(e.Name.MustValue().Position, e.Name.MustValue())
		}
		return param
	})
	paramTypes := stlslices.Map(params, func(_ int, p *local.Param) types.Type {
		return p.Type()
	})

	var ret types.Type = types.NoThing
	if retNode, ok := node.Ret.Value(); ok {
		ret = self.analyseType(retNode, append([]typeAnalyser{NewNoReturnTypeAnalyser()}, analysers...)...)
	}
	return global.NewFuncDecl(types.NewFuncType(ret, paramTypes...), name, params...)
}
