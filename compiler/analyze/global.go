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

func (a *Analyzer) analyzeTypeDef(global *ast.TypeDef) {
	decl, _ := a.scope.LookupType(global.Name.OriginText)

	a.typedefStack.Add(decl)
	defer a.typedefStack.Remove(decl)

	underlying := a.analyzeType(global.Type)
	decl.Type = types.NewCustomType(global.Name.OriginText, underlying)
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
