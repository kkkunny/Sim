package analyse

import (
	"github.com/kkkunny/stl/container/either"
	"github.com/kkkunny/stl/container/set"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/ast"
	errors "github.com/kkkunny/Sim/compiler/error"
	"github.com/kkkunny/Sim/compiler/hir/global"
	"github.com/kkkunny/Sim/compiler/hir/local"
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/hir/values"
	"github.com/kkkunny/Sim/compiler/reader"
)

func (self *Analyser) analyseParam(node ast.Param, analysers ...typeAnalyser) *local.Param {
	mut := !node.Mutable.IsNone()
	name := stlval.TernaryAction(node.Name.IsSome(), func() string {
		return node.Name.MustValue().Source()
	}, func() string {
		return ""
	})
	typ := self.analyseType(node.Type, analysers...)
	return local.NewParam(mut, name, typ)
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
		ret = self.analyseType(retNode, append(analysers, self.noReturnTypeAnalyser())...)
	}
	return global.NewFuncDecl(types.NewFuncType(ret, paramTypes...), name, params...)
}

// 检查类型是否循环（是否不能确定size）
func (self *Analyser) checkTypeCircle(trace set.Set[types.Type], t types.Type) bool {
	if trace.Contain(t) {
		return true
	}
	trace.Add(t)
	defer func() {
		trace.Remove(t)
	}()

	switch typ := t.(type) {
	case types.NoThingType, types.NoReturnType, types.NumType, types.BoolType, types.StrType, types.RefType, types.CallableType:
		return false
	case types.ArrayType:
		return self.checkTypeCircle(trace, typ.Elem())
	case types.TupleType:
		for _, e := range typ.Elems() {
			if self.checkTypeCircle(trace, e) {
				return true
			}
		}
		return false
	case types.StructType:
		for _, f := range typ.Fields().Values() {
			if self.checkTypeCircle(trace, f.Type()) {
				return true
			}
		}
		return false
	case types.EnumType:
		for iter := typ.EnumFields().Iterator(); iter.Next(); {
			if e, ok := iter.Value().E2().Elem(); ok {
				if self.checkTypeCircle(trace, e) {
					return true
				}
			}
		}
		return false
	case global.TypeDef:
		return self.checkTypeCircle(trace, typ.Target())
	default:
		panic("unreachable")
	}
}

func (self *Analyser) buildinPkg() *global.Package {
	if self.pkg.IsBuildIn() {
		return self.pkg
	}
	return stlval.IgnoreWith(stlslices.FindFirst(self.pkg.GetLinkedPackages(), func(_ int, pkg *global.Package) bool {
		return pkg.IsBuildIn()
	}))
}

// 类型是否有默认值
func (self *Analyser) hasTypeDefault(typ types.Type) bool {
	switch t := typ.(type) {
	case types.NoThingType, types.NoReturnType, types.RefType:
		return false
	case types.NumType, types.BoolType, types.StrType:
		return true
	case types.CallableType:
		if types.Is[types.NoThingType](t.Ret(), true) {
			return true
		}
		return self.hasTypeDefault(t.Ret())
	case types.ArrayType:
		return self.hasTypeDefault(t.Elem())
	case types.TupleType:
		return stlslices.All(t.Elems(), func(_ int, e types.Type) bool {
			return self.hasTypeDefault(e)
		})
	case types.StructType:
		return stlslices.All(t.Fields().Values(), func(_ int, f *types.Field) bool {
			return self.hasTypeDefault(f.Type())
		})
	case types.EnumType:
		return stlslices.All(t.EnumFields().Values(), func(_ int, f *types.EnumField) bool {
			e, ok := f.Elem()
			if !ok {
				return true
			}
			return self.hasTypeDefault(e)
		})
	case types.CustomType:
		return self.hasTypeDefault(t.Target())
	case types.AliasType:
		return self.hasTypeDefault(t.Target())
	default:
		panic("unreachable")
	}
}

func (self *Analyser) tryAnalyseIdent(node *ast.Ident) (either.Either[types.Type, values.Value], bool) {
	t, ok := self.tryAnalyseIdentType((*ast.IdentType)(node))
	if ok {
		return either.Left[types.Type, values.Value](t), true
	}
	v, ok := self.tryAnalyseIdentExpr((*ast.IdentExpr)(node))
	if ok {
		return either.Right[types.Type, values.Value](v), true
	}
	return stlval.Default[either.Either[types.Type, values.Value]](), false
}

// 获取类型默认值
func (self *Analyser) getTypeDefaultValue(pos reader.Position, t types.Type) *local.DefaultExpr {
	if !self.hasTypeDefault(t) {
		errors.ThrowCanNotGetDefault(pos, t)
	}
	return local.NewDefaultExpr(t)
}
