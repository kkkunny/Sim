package analyse

import (
	"github.com/kkkunny/stl/container/either"
	"github.com/kkkunny/stl/container/linkedhashmap"
	"github.com/kkkunny/stl/container/optional"
	"github.com/kkkunny/stl/container/set"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/config"
	errors "github.com/kkkunny/Sim/compiler/error"
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/global"
	"github.com/kkkunny/Sim/compiler/hir/local"
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/reader"
)

// 当前文件作用域
func (self *Analyser) getFileByPath(pos reader.Position) *hir.File {
	return self.allFiles.Get(pos.Reader.Path())
}

// buildin包
func (self *Analyser) buildinPkg() *hir.Package {
	return self.allPkgs.Get(config.BuildInPkgPath)
}

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
	paramTypes := stlslices.Map(params, func(_ int, p *local.Param) hir.Type {
		return p.Type()
	})

	var ret hir.Type = types.NoThing
	if retNode, ok := node.Ret.Value(); ok {
		ret = self.analyseType(retNode, append(analysers, self.noReturnTypeAnalyser())...)
	}
	return global.NewFuncDecl(types.NewFuncType(ret, paramTypes...), name, params...)
}

// 检查类型是否循环（是否不能确定size）
func (self *Analyser) checkTypeCircle(trace set.Set[hir.Type], t hir.Type) bool {
	if trace.Contain(t) {
		return true
	}
	trace.Add(t)
	defer func() {
		trace.Remove(t)
	}()

	switch typ := t.(type) {
	case types.NoThingType, types.NoReturnType, types.NumType, types.BoolType, types.StrType, types.RefType, types.CallableType, types.GenericParamType:
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
	case types.TypeDef:
		return self.checkTypeCircle(trace, typ.Target())
	default:
		panic("unreachable")
	}
}

// 类型是否有默认值
func (self *Analyser) hasTypeDefault(typ hir.Type) bool {
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
		return stlslices.All(t.Elems(), func(_ int, e hir.Type) bool {
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

func (self *Analyser) tryAnalyseIdent(node *ast.Ident, typeAnalysers ...typeAnalyser) (res either.Either[hir.Type, hir.Value], ok bool) {
	scope := self.scope
	if pkgToken, ok := node.Pkg.Value(); ok {
		scope, ok = self.getFileByPath(node.Position()).GetExternPackage(pkgToken.Source())
		if !ok {
			errors.ThrowUnknownIdentifierError(pkgToken.Position, pkgToken)
		}
	}

	name := node.Name.Source()
	defer func() {
		if !ok {
			if self.scope.Package().IsBuildIn() {
				var buildinType hir.Type
				switch name {
				case "__buildin_isize":
					buildinType = types.Isize
				case "__buildin_i8":
					buildinType = types.I8
				case "__buildin_i16":
					buildinType = types.I16
				case "__buildin_i32":
					buildinType = types.I32
				case "__buildin_i64":
					buildinType = types.I64
				case "__buildin_usize":
					buildinType = types.Usize
				case "__buildin_u8":
					buildinType = types.U8
				case "__buildin_u16":
					buildinType = types.U16
				case "__buildin_u32":
					buildinType = types.U32
				case "__buildin_u64":
					buildinType = types.U64
				case "__buildin_f16":
					buildinType = types.F16
				case "__buildin_f32":
					buildinType = types.F32
				case "__buildin_f64":
					buildinType = types.F64
				case "__buildin_f128":
					buildinType = types.F128
				case "__buildin_bool":
					buildinType = types.Bool
				case "__buildin_str":
					buildinType = types.Str
				}
				if buildinType != nil {
					res = either.Left[hir.Type, hir.Value](buildinType)
					ok = true
				}
			}
		}
	}()

	// TODO: 外部包不能再次allowLinked
	identObj, ok := scope.GetIdent(name, true)
	if !ok {
		return stlval.Default[either.Either[hir.Type, hir.Value]](), false
	}
	if g, ok := identObj.(hir.Global); ok && !g.Package().Equal(self.scope.Package()) && !g.Public() {
		return stlval.Default[either.Either[hir.Type, hir.Value]](), false
	}

	switch ident := identObj.(type) {
	case *global.FuncDef:
		genericArgs := self.analyseOptionalGenericArgList(ident.GenericParams(), node.Position(), node.GenericArgs)
		if len(genericArgs) == 0 {
			return either.Right[hir.Type, hir.Value](ident), true
		}
		return either.Right[hir.Type, hir.Value](local.NewGenericFuncInstExpr(ident, genericArgs...)), true
	case hir.Value:
		return either.Right[hir.Type, hir.Value](ident), true
	case global.CustomTypeDef:
		genericArgs := self.analyseOptionalGenericArgList(ident.GenericParams(), node.Position(), node.GenericArgs, typeAnalysers...)
		if len(genericArgs) == 0 {
			return either.Left[hir.Type, hir.Value](ident), true
		}
		return either.Left[hir.Type, hir.Value](global.NewGenericCustomTypeDef(ident, genericArgs...)), true
	case hir.Type:
		return either.Left[hir.Type, hir.Value](ident), true
	}
	return stlval.Default[either.Either[hir.Type, hir.Value]](), false
}

func (self *Analyser) analyseTraitIdent(node *ast.IdentType) *global.Trait {
	scope := self.scope
	if pkgToken, ok := node.Pkg.Value(); ok {
		scope, ok = self.getFileByPath(node.Position()).GetExternPackage(pkgToken.Source())
		if !ok {
			errors.ThrowUnknownIdentifierError(pkgToken.Position, pkgToken)
		}
	}

	// TODO: 外部包不能再次allowLinked
	identObj, ok := scope.GetIdent(node.Name.Source(), true)
	if !ok {
		errors.ThrowUnknownIdentifierError(node.Name.Position, node.Name)
	}
	trait, ok := identObj.(*global.Trait)
	if !ok {
		errors.ThrowExpectTraitError(node.Position())
	} else if !trait.Package().Equal(self.scope.Package()) && !trait.Public() {
		errors.ThrowUnknownIdentifierError(node.Name.Position, node.Name)
	}

	return trait
}

// 获取类型默认值
func (self *Analyser) getTypeDefaultValue(pos reader.Position, t hir.Type) *local.DefaultExpr {
	if !self.hasTypeDefault(t) {
		errors.ThrowCanNotGetDefault(pos, t)
	}
	return local.NewDefaultExpr(t)
}

func (self *Analyser) analyseGenericParamsList(node optional.Optional[*ast.GenericParamList]) linkedhashmap.LinkedHashMap[string, types.GenericParamType] {
	res := linkedhashmap.StdWith[string, types.GenericParamType]()
	if genericParamsNode, ok := node.Value(); ok {
		for _, genericParamNode := range genericParamsNode.Params {
			genericParamName := genericParamNode.Name.Source()
			if res.Contain(genericParamName) {
				errors.ThrowIdentifierDuplicationError(genericParamNode.Name.Position, genericParamNode.Name)
			}
			if restraintNode, ok := genericParamNode.Restraint.Value(); ok {
				restraint := self.analyseTraitIdent(restraintNode)
				res.Set(genericParamName, types.NewGenericParam(genericParamName, restraint))
			} else {
				res.Set(genericParamName, types.NewGenericParam(genericParamName))
			}
		}
	}
	return res
}

func (self *Analyser) analyseOptionalGenericArgList(genericParams []types.GenericParamType, pos reader.Position, node optional.Optional[*ast.GenericArgList], typeAnalysers ...typeAnalyser) []hir.Type {
	genericArgsNode, ok := node.Value()
	if !ok {
		return nil
	}
	if len(genericParams) != len(genericArgsNode.Args) {
		errors.ThrowParameterNumberNotMatchError(pos, uint(len(genericParams)), uint(len(genericArgsNode.Args)))
	}
	if len(genericParams) == 0 {
		return nil
	}
	return stlslices.Map(genericArgsNode.Args, func(i int, genericArgNode ast.Type) hir.Type {
		t := self.analyseType(genericArgNode, typeAnalysers...)
		restraint, ok := genericParams[i].Restraint()
		if !ok {
			return t
		}
		if trait := restraint.(*global.Trait); !trait.HasBeImpled(t) {
			errors.ThrowNotImplTrait(genericArgNode.Position(), t, trait)
		}
		return t
	})
}
