package analyze

import (
	"path/filepath"

	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/config"
	"github.com/kkkunny/Sim/compiler/hir/scopes"
	"github.com/kkkunny/Sim/compiler/hir/stmts"
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/report"
	"github.com/kkkunny/Sim/compiler/token"
)

func (a *Analyzer) analyzeImport(global *ast.Import) error {
	lastPkgToken := stlslices.Last(global.Pkgs)
	name := lastPkgToken.OriginText

	_, ok := a.scope.LookupPkg(name)
	if ok {
		a.reporter.Fatalf(
			lastPkgToken.Position,
			report.Errors.RepeatedIdentifier,
			name,
		)
	}

	paths := stlslices.Map(global.Pkgs, func(_ int, tok token.Token) string {
		return tok.OriginText
	})
	dirpath := filepath.Join(append([]string{config.StdPkgPath}, paths...)...)

	if _, ok = a.pkgScopes[dirpath]; !ok {
		_, err := analyzeDir(dirpath, a.reporter, a)
		if err != nil {
			return err
		}
	}

	ir, scope := a.pkgScopes[dirpath].Unpack()
	a.scope.Package().AddExternal(name, scope)
	a.ir.Dependencies = append(a.ir.Dependencies, ir)
	return nil
}

func (a *Analyzer) analyzeTypeDecl(global ast.Global) *stmts.TypeDef {
	switch global := global.(type) {
	case *ast.Let, *ast.Import:
		return nil
	case *ast.TypeDef:
		return a.analyzeCustomTypeDecl(global)
	default:
		panic("unreachable")
	}
}

func (a *Analyzer) analyzeCustomTypeDecl(global *ast.TypeDef) *stmts.TypeDef {
	_, ok := a.scope.LookupType(global.Name.OriginText)
	if ok {
		a.reporter.Fatalf(
			global.Name.Position,
			report.Errors.RepeatedIdentifier,
			global.Name.OriginText,
		)
		return nil
	}

	decl := stmts.NewTypeDef(global.Public)
	a.scope.(*scopes.PkgScope).AddType(global.Name.OriginText, decl)
	a.typedefAsts[decl] = global
	return decl
}

func (a *Analyzer) analyzeTypeDef(global ast.Global) {
	switch global := global.(type) {
	case *ast.Let, *ast.Import:
		return
	case *ast.TypeDef:
		a.analyzeCustomTypeDef(global)
	default:
		panic("unreachable")
	}
}

func (a *Analyzer) analyzeCustomTypeDef(global *ast.TypeDef) types.Type {
	name := global.Name.OriginText
	decl, _ := a.scope.LookupType(name)
	if decl.Type != nil {
		return decl.Type
	}

	var ct types.CustomType
	var setter func(types.Type)
	switch t := global.Type.(type) {
	case *ast.IdentType:
		if _, ok := a.scope.LookupType(t.Name.OriginText); ok {
			a.reporter.Fatalf(
				global.Name.Position,
				report.Errors.InvalidRecursionType,
			)
		}
		switch a.analyzeBuildInIdentType(t).(type) {
		case types.SintType:
			ctt, s := types.DelayNewCustomType[types.SintType](name)
			ct, setter = ctt, func(t types.Type) { s(t.(types.SintType)) }
		case types.UintType:
			ctt, s := types.DelayNewCustomType[types.UintType](name)
			ct, setter = ctt, func(t types.Type) { s(t.(types.UintType)) }
		case types.FloatType:
			ctt, s := types.DelayNewCustomType[types.FloatType](name)
			ct, setter = ctt, func(t types.Type) { s(t.(types.FloatType)) }
		case types.BooleanType:
			ctt, s := types.DelayNewCustomType[types.BooleanType](name)
			ct, setter = ctt, func(t types.Type) { s(t.(types.BooleanType)) }
		default:
			panic("unreachable")
		}
	case *ast.FuncType:
		ctt, s := types.DelayNewCustomType[types.FuncType](name)
		ct, setter = ctt, func(t types.Type) { s(t.(types.FuncType)) }
	case *ast.TupleType:
		ctt, s := types.DelayNewCustomType[types.TupleType](name)
		ct, setter = ctt, func(t types.Type) { s(t.(types.TupleType)) }
	case *ast.ArrayType:
		ctt, s := types.DelayNewCustomType[types.ArrayType](name)
		ct, setter = ctt, func(t types.Type) { s(t.(types.ArrayType)) }
	case *ast.UnionType:
		ctt, s := types.DelayNewCustomType[types.UnionType](name)
		ct, setter = ctt, func(t types.Type) { s(t.(types.UnionType)) }
	case *ast.RefType:
		ctt, s := types.DelayNewCustomType[types.RefType](name)
		ct, setter = ctt, func(t types.Type) { s(t.(types.RefType)) }
	case *ast.StructType:
		ctt, s := types.DelayNewCustomType[types.StructType](name)
		ct, setter = ctt, func(t types.Type) { s(t.(types.StructType)) }
	}
	decl.Type = ct
	setter(a.analyzeType(global.Type))

	if types.CheckRecursion(ct) {
		a.reporter.Fatalf(
			global.Name.Position,
			report.Errors.InvalidRecursionType,
		)
	}

	return ct
}

func (a *Analyzer) analyzeGlobalValueDecl(global ast.Global) {
	switch global := global.(type) {
	case *ast.TypeDef, *ast.Import:
		return
	case *ast.Let:
		a.analyzeGlobalLetDecl(global)
	default:
		panic("unreachable")
	}
}

func (a *Analyzer) analyzeGlobalLetDecl(global *ast.Let) {
	_, ok := a.scope.LookupValue(global.Name.OriginText)
	if ok {
		a.reporter.Fatalf(
			global.Name.Position,
			report.Errors.RepeatedIdentifier,
			global.Name.OriginText,
		)
	}

	var t types.Type
	if tAst, ok := global.Type.Value(); ok {
		t = a.analyzeType(tAst)
	} else {
		v, ok := global.Value.MustValue().(*ast.Func)
		if !ok {
			// TODO: 非函数定义的全局变量的声明解析
			if global.Name.OriginText == "main" {
				a.reporter.Fatalf(
					global.Name.Position,
					report.Errors.InvalidMainFunction,
				)
			}
			return
		}
		t = a.analyzeFuncDecl(v)
	}

	if global.Name.OriginText == "main" {
		expectType := types.NewFuncType(types.Unit)
		if !t.Equal(expectType) {
			a.reporter.Fatalf(
				global.Name.Position,
				report.Errors.UnexpectedExpression,
				expectType, t,
			)
		}
	}

	let := &stmts.Let{
		Pub:    global.Public,
		Global: true,
		Mut:    global.Mut,
		Type:   t,
		Name:   global.Name.OriginText,
	}

	for _, attrAst := range global.Attributes {
		switch attrAst := attrAst.(type) {
		case *ast.Extern:
			let.ExternalName = optional.Some(attrAst.Name.OriginText)
		default:
			panic("unreachable")
		}
	}

	a.scope.AddValue(let)
}

func (a *Analyzer) analyzeGlobalValueDef(global ast.Global) stmts.Global {
	switch global := global.(type) {
	case *ast.TypeDef, *ast.Import:
		return nil
	case *ast.Let:
		return a.analyzeLetDef(global, true)
	default:
		panic("unreachable")
	}
}
