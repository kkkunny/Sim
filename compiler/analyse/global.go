package analyse

import (
	"github.com/kkkunny/stl/container/linkedhashmap"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlos "github.com/kkkunny/stl/os"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/config"
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/global"
	"github.com/kkkunny/Sim/compiler/hir/local"
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/reader"
	"github.com/kkkunny/Sim/compiler/util"

	errors "github.com/kkkunny/Sim/compiler/error"

	"github.com/kkkunny/Sim/compiler/token"
)

func (self *Analyser) analyseImport(node *ast.Import) *hir.Package {
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
			errors.ThrowPackageCircularReference(stlslices.Last(node.Paths).Position, stlslices.Map(e.chain, func(i int, pkg *hir.Package) stlos.FilePath {
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
	return self.pkg.AppendGlobal(node.Public, global.NewTrait(name)).(*global.Trait)
}

func (self *Analyser) declTypeDef(node *ast.TypeDef) global.CustomTypeDef {
	name := node.Name.Source()
	_, exist := self.pkg.GetIdent(name)
	if exist {
		errors.ThrowIdentifierDuplicationError(node.Name.Position, node.Name)
	}

	// 编译参数
	compileParams := linkedhashmap.StdWith[string, types.GenericParamType]()
	if genericParamsNode, ok := node.GenericParams.Value(); ok {
		for _, genericParamNode := range genericParamsNode.Params {
			genericParamName := genericParamNode.Source()
			if compileParams.Contain(genericParamName) {
				errors.ThrowIdentifierDuplicationError(genericParamNode.Position, genericParamNode)
			}
			compileParams.Set(genericParamName, types.NewGenericParam(genericParamName))
		}
	}

	return self.pkg.AppendGlobal(node.Public, global.NewCustomTypeDef(name, compileParams.Values(), types.NoThing)).(global.CustomTypeDef)
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
		if !decl.AddMethod(method) {
			errors.ThrowIdentifierDuplicationError(methodNode.Name.Position, methodNode.Name)
		}
	}
	return decl
}

func (self *Analyser) defTypeDef(node *ast.TypeDef) global.CustomTypeDef {
	decl := stlval.IgnoreWith(self.pkg.GetIdent(node.Name.Source())).(global.CustomTypeDef)
	fnDeclCompileParamAnalyser := stlslices.Map(decl.GenericParams(), func(_ int, compileParam types.GenericParamType) typeAnalyser {
		return self.genericParamAnalyserWith(compileParam, true)
	})
	decl.SetTarget(self.analyseType(node.Target, append([]typeAnalyser{self.structTypeAnalyser(), self.enumTypeAnalyser(), self.selfTypeAnalyserWith(decl, true)}, fnDeclCompileParamAnalyser...)...))
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

	// 泛型参数
	genericParams := linkedhashmap.StdWith[string, types.GenericParamType]()
	if genericParamsNode, ok := node.GenericParams.Value(); ok {
		for _, genericParamNode := range genericParamsNode.Params {
			genericParamName := genericParamNode.Source()
			if genericParams.Contain(genericParamName) {
				errors.ThrowIdentifierDuplicationError(genericParamNode.Position, genericParamNode)
			}
			genericParams.Set(genericParamName, types.NewGenericParam(genericParamName))
		}
	}
	fnDeclGenericParamAnalyser := stlslices.Map(genericParams.Values(), func(_ int, compileParam types.GenericParamType) typeAnalyser {
		return self.genericParamAnalyserWith(compileParam, true)
	})

	f := global.NewFuncDef(self.analyseFuncDecl(node.FuncDecl, fnDeclGenericParamAnalyser...), genericParams.Values(), attrs...)
	if stlval.IgnoreWith(f.GetName()) == "main" {
		mainType := types.NewFuncType(types.NoThing)
		if !f.Type().Equal(types.NewFuncType(types.NoThing)) {
			errors.ThrowTypeMismatchError(node.Position(), f.Type(), mainType)
		}
	}
	return self.pkg.AppendGlobal(node.Public, f).(*global.FuncDef)
}

func (self *Analyser) declMethodDef(node *ast.FuncDef) global.MethodDef {
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

	// 泛型参数
	genericParams := linkedhashmap.StdWith[string, types.GenericParamType]()
	for _, genericParamNode := range customType.GenericParams() {
		genericParams.Set(genericParamNode.String(), genericParamNode)
	}
	fnDeclGenericParamAnalyser := stlslices.Map(genericParams.Values(), func(_ int, compileParam types.GenericParamType) typeAnalyser {
		return self.genericParamAnalyserWith(compileParam, true)
	})
	genericParams = linkedhashmap.StdWith[string, types.GenericParamType]()
	if genericParamsNode, ok := node.GenericParams.Value(); ok {
		for _, genericParamNode := range genericParamsNode.Params {
			genericParamName := genericParamNode.Source()
			if genericParams.Contain(genericParamName) {
				errors.ThrowIdentifierDuplicationError(genericParamNode.Position, genericParamNode)
			}
			compileParam := types.NewGenericParam(genericParamName)
			genericParams.Set(genericParamName, compileParam)
			fnDeclGenericParamAnalyser = append(fnDeclGenericParamAnalyser, self.genericParamAnalyserWith(compileParam, true))
		}
	}

	f := global.NewOriginMethodDef(
		customType,
		self.analyseFuncDecl(node.FuncDecl, append([]typeAnalyser{self.selfTypeAnalyserWith(customType, true)}, fnDeclGenericParamAnalyser...)...),
		genericParams.Values(),
		attrs...,
	)
	if !customType.AddMethod(f) {
		errors.ThrowIdentifierDuplicationError(node.Position(), node.Name)
	}

	return self.pkg.AppendGlobal(node.Public, f).(global.MethodDef)
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

	v := global.NewVarDef(node.Var.Mutable, name, self.analyseType(node.Var.Type.MustValue()), attrs...)
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

func (self *Analyser) defMethodDef(node *ast.FuncDef) global.MethodDef {
	ct := stlval.IgnoreWith(self.pkg.GetIdent(node.SelfType.MustValue().Source(), false)).(global.CustomTypeDef)
	decl := stlval.IgnoreWith(ct.GetMethod(node.Name.Source())).(*global.OriginMethodDef)
	if node.Body.IsNone() {
		return decl
	}

	decl.SetBody(self.analyseFuncBody(decl, node.Params, node.Body.MustValue()))
	return decl
}

func (self *Analyser) defSingleGlobalVariable(node *ast.SingleVariableDef) *global.VarDef {
	decl := stlval.IgnoreWith(self.pkg.GetIdent(node.Var.Name.Source())).(*global.VarDef)

	var linkname = ""
	for _, attr := range decl.Attrs() {
		switch attr := attr.(type) {
		case *global.VarAttrLinkName:
			linkname = attr.Name()
		}
	}

	if valueNode, ok := node.Value.Value(); ok {
		decl.SetValue(self.expectExpr(decl.Type(), valueNode))
	} else if linkname == "" {
		decl.SetValue(self.getTypeDefaultValue(node.Var.Type.MustValue().Position(), decl.Type()))
	}
	return decl
}

func (self *Analyser) defMultiGlobalVariable(node *ast.MultipleVariableDef) []*global.VarDef {
	decls := stlslices.Map(node.Vars, func(_ int, item ast.VarDef) *global.VarDef {
		return stlval.IgnoreWith(self.pkg.GetIdent(item.Name.Source())).(*global.VarDef)
	})
	varTypes := stlslices.Map(decls, func(_ int, decl *global.VarDef) hir.Type {
		return decl.Type()
	})

	var value hir.Value
	if valueNode, ok := node.Value.Value(); ok {
		value = self.expectExpr(types.NewTupleType(varTypes...), valueNode)
	} else {
		elems := make([]hir.Value, len(decls))
		for i, varDef := range node.Vars {
			elems[i] = self.getTypeDefaultValue(varDef.Type.MustValue().Position(), varTypes[i])
		}
		value = local.NewTupleExpr(elems...)
	}

	if tuple, ok := value.(*local.TupleExpr); ok {
		for i, v := range tuple.Elems() {
			decls[i].SetValue(v)
		}
		return decls
	}

	v := global.NewVarDef(false, "", value.Type())
	v.SetValue(value)
	for i, decl := range decls {
		decl.SetValue(local.NewExtractExpr(v, uint(i)))
	}
	return decls
}
