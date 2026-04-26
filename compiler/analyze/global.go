package analyze

import (
	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/hir/scopes"
	"github.com/kkkunny/Sim/compiler/hir/stmts"
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/report"
)

func (a *Analyzer) analyzeTypeDecl(global *ast.TypeDef) *stmts.TypeDef {
	_, ok := a.scope.LookupType(global.Name.OriginText)
	if ok {
		a.reporter.Fatalf(
			global.Name.Position,
			report.Errors.RepeatedIdentifier,
			global.Name.OriginText,
		)
		return nil
	}

	decl := stmts.NewTypeDef()
	a.scope.(*scopes.PkgScope).AddType(global.Name.OriginText, decl)
	a.typedefAsts[decl] = global
	return decl
}

func (a *Analyzer) analyzeTypeDef(t *ast.IdentType) types.Type {
	name := t.Name.OriginText
	decl, _ := a.scope.LookupType(name)
	if decl.Type != nil {
		return decl.Type
	}

	tdAst := a.typedefAsts[decl]

	var ct types.CustomType
	var setter func(types.Type)
	switch t := tdAst.Type.(type) {
	case *ast.IdentType:
		if _, ok := a.scope.LookupType(t.Name.OriginText); ok {
			a.reporter.Fatalf(
				tdAst.Name.Position,
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
	}
	decl.Type = ct
	setter(a.analyzeType(tdAst.Type))

	if types.CheckRecursion(ct) {
		a.reporter.Fatalf(
			tdAst.Name.Position,
			report.Errors.InvalidRecursionType,
		)
	}

	return ct
}

func (a *Analyzer) analyzeGlobalDecl(global ast.Global) {
	switch global := global.(type) {
	case *ast.Let:
		a.analyzeGlobalLetDecl(global)
	}
}

func (a *Analyzer) analyzeGlobalLetDecl(global *ast.Let) {
	_, ok := a.scope.Lookup(global.Name.OriginText)
	if ok {
		a.reporter.Fatalf(
			global.Name.Position,
			report.Errors.RepeatedIdentifier,
			global.Name.OriginText,
		)
	}

	if global.Name.OriginText == "main" && global.Mut {
		a.reporter.Fatalf(
			global.Name.Position,
			report.Errors.MustImmutable,
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

	a.scope.AddValue(&stmts.Let{
		Global: true,
		Mut:    global.Mut,
		Type:   t,
		Name:   global.Name.OriginText,
	})
}

func (a *Analyzer) analyzeGlobalDef(global ast.Global) stmts.Global {
	switch global := global.(type) {
	case *ast.Let:
		return a.analyzeLetDef(global, true)
	default:
		panic("unreachable")
	}
}
