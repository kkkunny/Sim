package analyze

import (
	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/hir/scopes"
	"github.com/kkkunny/Sim/compiler/hir/stmts"
	"github.com/kkkunny/Sim/compiler/report"
)

func (a *Analyzer) analyzeGlobal(global ast.Global) stmts.Global {
	switch global := global.(type) {
	case *ast.Let:
		return a.analyzeLet(global, true)
	case *ast.TypeDef:
		return a.analyzeTypeDef(global)
	default:
		panic("unreachable")
	}
}

func (a *Analyzer) analyzeTypeDef(global *ast.TypeDef) *stmts.TypeDef {
	_, ok := a.scope.LookupType(global.Name.OriginText)
	if ok {
		a.reporter.Fatalf(
			global.Name.Position,
			report.Errors.RepeatedIdentifier,
			global.Name.OriginText,
		)
		return nil
	}

	underlying := a.analyzeType(global.Type)
	td := stmts.NewTypeDef(global.Name.OriginText, underlying)
	a.scope.(*scopes.PkgScope).AddType(global.Name.OriginText, td.Type)
	return td
}
