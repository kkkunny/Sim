package analyse

import (
	"github.com/kkkunny/stl/container/hashmap"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlos "github.com/kkkunny/stl/os"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/config"
	"github.com/kkkunny/Sim/compiler/hir/global"
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/hir/values"
	"github.com/kkkunny/Sim/compiler/reader"
	"github.com/kkkunny/Sim/compiler/util"

	errors "github.com/kkkunny/Sim/compiler/error"

	"github.com/kkkunny/Sim/compiler/token"
)

func (self *Analyser) analyseImport(node *ast.Import) *global.Package {
	// 包名
	var pkgName string
	var importAll bool
	if alias, ok := node.Alias.Value(); ok && alias.Is(token.IDENT) {
		pkgName = alias.Source()
	} else {
		importAll = alias.Is(token.MUL)
		pkgName = stlslices.Last(node.Paths).Source()
	}

	// 包地址
	paths := stlslices.Map(node.Paths, func(_ int, v token.Token) string {
		return v.Source()
	})
	pkgPath := config.OfficialPkgPath.Join(paths...)

	// 导入包
	pkg, err := self.importPackage(pkgPath, pkgName, importAll)
	if err != nil {
		switch e := err.(type) {
		case *importPackageCircularError:
			errors.ThrowPackageCircularReference(stlslices.Last(node.Paths).Position, stlslices.Map(e.chain, func(i int, pkg *global.Package) stlos.FilePath {
				return pkg.Path()
			}))
		case *importPackageDuplicationError:
			errors.ThrowIdentifierDuplicationError(stlslices.Last(node.Paths).Position, stlslices.Last(node.Paths))
		case *importPackageInvalidError:
			errors.ThrowInvalidPackage(reader.MixPosition(stlslices.First(node.Paths).Position, stlslices.Last(node.Paths).Position), node.Paths)
		default:
			panic("unreachable")
		}
	}
	return pkg
}

func (self *Analyser) declTrait(node *ast.Trait) *global.Trait {
	name := node.Name.Source()
	_, exist := self.pkg.GetIdent(name)
	if exist {
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
	}
	return self.pkg.AppendGlobal(node.Public, global.NewTrait(name, hashmap.StdWithCap[string, *global.FuncDecl](uint(len(node.Methods))))).(*global.Trait)
}

func (self *Analyser) declTypeDef(node *ast.TypeDef) global.CustomTypeDef {
	name := node.Name.Source()
	_, exist := self.pkg.GetIdent(name)
	if exist {
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
	}
	return self.pkg.AppendGlobal(node.Public, global.NewCustomTypeDef(name, types.NoThing)).(global.CustomTypeDef)
}

func (self *Analyser) declTypeAlias(node *ast.TypeAlias) global.AliasTypeDef {
	name := node.Name.Source()
	_, exist := self.pkg.GetIdent(name)
	if exist {
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
	}
	return self.pkg.AppendGlobal(node.Public, global.NewAliasTypeDef(name, types.NoThing)).(global.AliasTypeDef)
}

func (self *Analyser) defTrait(node *ast.Trait) *global.Trait {
	decl := stlval.IgnoreWith(self.pkg.GetIdent(node.Name.Source())).(*global.Trait)

	for _, methodNode := range node.Methods {
		method := self.analyseFuncDecl(*methodNode, self.selfTypeAnalyser(true))
		if decl.Methods.Contain(method.Name()) {
			errors.ThrowIdentifierDuplicationError(methodNode.Name.Position, methodNode.Name)
		}
		decl.Methods.Set(method.Name(), method)
	}
	return decl
}

func (self *Analyser) defTypeDef(node *ast.TypeDef) global.CustomTypeDef {
	decl := stlval.IgnoreWith(self.pkg.GetIdent(node.Name.Source())).(global.CustomTypeDef)
	decl.SetTarget(self.analyseType(node.Target, self.structTypeAnalyser(), self.enumTypeAnalyser(), self.selfTypeAnalyserWith(decl, true)))
	return decl
}

func (self *Analyser) defTypeAlias(node *ast.TypeAlias) global.AliasTypeDef {
	decl := stlval.IgnoreWith(self.pkg.GetIdent(node.Name.Source())).(global.AliasTypeDef)
	decl.SetTarget(self.analyseType(node.Target))
	return decl
}

func (self *Analyser) declFuncDef(node *ast.FuncDef) *global.FuncDef {
	name := node.Name.Source()
	_, exist := self.pkg.GetIdent(name, false)
	if exist {
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
	}

	var attrs []global.FuncAttr
	var linkName string
	for _, attrObj := range node.Attrs {
		switch attr := attrObj.(type) {
		case *ast.Extern:
			temp := attr.Name.Source()
			linkName = util.ParseEscapeCharacter(temp[1:len(temp)-1], `\"`, `"`)
			attrs = append(attrs, global.WithLinkNameFuncAttr(linkName))
		case *ast.Inline:
			attrs = append(attrs, global.WithInlineFuncAttr(true))
		case *ast.NoInline:
			attrs = append(attrs, global.WithInlineFuncAttr(false))
		case *ast.VarArg:
			attrs = append(attrs, global.WithVarargFuncAttr())
		default:
			panic("unreachable")
		}
	}
	if node.Body.IsNone() && linkName == "" {
		errors.ThrowExpectAttribute(node.Position(), new(ast.Extern))
	}

	f := global.NewFuncDef(self.analyseFuncDecl(node.FuncDecl), attrs...)
	if f.Name() == "main" {
		mainType := types.NewFuncType(types.NoThing)
		if !f.Type().Equal(types.NewFuncType(types.NoThing)) {
			errors.ThrowTypeMismatchErrorV2(node.Position(), f.Type(), mainType)
		}
	}
	return self.pkg.AppendGlobal(node.Public, f).(*global.FuncDef)
}

func (self *Analyser) declMethodDef(node *ast.FuncDef) *global.MethodDef {
	customTypeObj, ok := self.pkg.GetIdent(node.SelfType.MustValue().Source(), false)
	if !ok || !stlval.Is[global.CustomTypeDef](customTypeObj) {
		errors.ThrowUnknownIdentifierError(node.SelfType.MustValue().Position, node.SelfType.MustValue())
	}
	customType := customTypeObj.(global.CustomTypeDef)

	var attrs []global.FuncAttr
	var linkName string
	for _, attrObj := range node.Attrs {
		switch attr := attrObj.(type) {
		case *ast.Extern:
			temp := attr.Name.Source()
			linkName = util.ParseEscapeCharacter(temp[1:len(temp)-1], `\"`, `"`)
			attrs = append(attrs, global.WithLinkNameFuncAttr(linkName))
		case *ast.Inline:
			attrs = append(attrs, global.WithInlineFuncAttr(true))
		case *ast.NoInline:
			attrs = append(attrs, global.WithInlineFuncAttr(false))
		case *ast.VarArg:
			attrs = append(attrs, global.WithVarargFuncAttr())
		default:
			panic("unreachable")
		}
	}
	if node.Body.IsNone() && linkName == "" {
		errors.ThrowExpectAttribute(node.Position(), new(ast.Extern))
	}

	f := global.NewMethodDef(customType, self.analyseFuncDecl(node.FuncDecl, self.selfTypeAnalyserWith(customType, true)), attrs...)
	if f.Name() == "main" {
		mainType := types.NewFuncType(types.NoThing)
		if !f.Type().Equal(types.NewFuncType(types.NoThing)) {
			errors.ThrowTypeMismatchErrorV2(node.Position(), f.Type(), mainType)
		}
	}

	if !customType.AddMethod(f) {
		errors.ThrowIdentifierDuplicationError(node.Position(), node.Name)
	}

	return self.pkg.AppendGlobal(node.Public, f).(*global.MethodDef)
}

func (self *Analyser) declSingleGlobalVariable(node *ast.SingleVariableDef) *global.VarDef {
	name := node.Var.Name.Source()
	_, exist := self.pkg.GetIdent(name, false)
	if exist {
		errors.ThrowIdentifierDuplicationError(node.Var.Name.Position, node.Var.Name)
	}

	var attrs []global.VarAttr
	for _, attrObj := range node.Attrs {
		switch attr := attrObj.(type) {
		case *ast.Extern:
			temp := attr.Name.Source()
			linkName := util.ParseEscapeCharacter(temp[1:len(temp)-1], `\"`, `"`)
			attrs = append(attrs, global.WithLinkNameVarAttr(linkName))
		default:
			panic("unreachable")
		}
	}

	v := global.NewVarDef(values.NewVarDecl(node.Var.Mutable, name, self.analyseType(node.Var.Type.MustValue())), attrs...)
	return self.pkg.AppendGlobal(node.Public, v).(*global.VarDef)
}

func (self *Analyser) declMultiGlobalVariable(node *ast.MultipleVariableDef) []*global.VarDef {
	return stlslices.Map(node.ToSingleList(), func(_ int, singleNode *ast.SingleVariableDef) *global.VarDef {
		return self.declSingleGlobalVariable(singleNode)
	})
}

func (self *Analyser) defFuncDef(node *ast.FuncDef) *global.FuncDef {
	decl := stlval.IgnoreWith(self.pkg.GetIdent(node.Name.Source())).(*global.FuncDef)
	if node.Body.IsNone() {
		return decl
	}

	decl.SetBody(self.analyseFuncBody(decl, node.Params, node.Body.MustValue()))
	return decl
}

func (self *Analyser) defMethodDef(node *ast.FuncDef) *global.MethodDef {
	ct := stlval.IgnoreWith(self.pkg.GetIdent(node.SelfType.MustValue().Source(), false)).(global.CustomTypeDef)
	decl := stlval.IgnoreWith(ct.GetMethod(node.Name.Source()))
	if node.Body.IsNone() {
		return decl
	}

	decl.SetBody(self.analyseFuncBody(decl, node.Params, node.Body.MustValue()))
	return decl
}

func (self *Analyser) defSingleGlobalVariable(node *ast.SingleVariableDef) *global.VarDef {
	decl := stlval.IgnoreWith(self.pkg.GetIdent(node.Var.Name.Source())).(*global.VarDef)

	// TODO: expr
	// if valueNode, ok := node.Value.Value(); ok {
	// 	decl.SetValue(self.expectExpr(v.Type, valueNode))
	// } else if v.ExternName == "" {
	// 	decl.SetValue(self.getTypeDefaultValue(node.Var.Type.MustValue().Position(), v.Type))
	// }
	return decl
}

func (self *Analyser) defMultiGlobalVariable(node *ast.MultipleVariableDef) []*global.VarDef {
	decls := stlslices.Map(node.Vars, func(_ int, item ast.VarDef) *global.VarDef {
		return stlval.IgnoreWith(self.pkg.GetIdent(item.Name.Source())).(*global.VarDef)
	})
	// varTypes := lo.Map(vars, func(item *oldhir.GlobalVarDef, _ int) oldhir.Type {
	// 	return item.GetType()
	// })

	// TODO: expr
	// var value oldhir.Expr
	// if valueNode, ok := node.Value.Value(); ok {
	// 	value = self.expectExpr(&oldhir.TupleType{Elems: varTypes}, valueNode)
	// } else {
	// 	tupleValue := &oldhir.Tuple{Elems: make([]oldhir.Expr, len(vars))}
	// 	for i, varDef := range node.Vars {
	// 		tupleValue.Elems[i] = self.getTypeDefaultValue(varDef.Type.MustValue().Position(), varTypes[i])
	// 	}
	// 	value = tupleValue
	// }
	return decls
}
