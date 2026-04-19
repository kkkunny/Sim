package analyze

import (
	"github.com/kkkunny/stl/container/optional"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/report"
)

func (a *Analyzer) analyzeBlock(block *ast.Block) *hir.Block {
	hirBlock := &hir.Block{}
	for _, stmt := range block.Stmts {
		hirBlock.Stmts = append(hirBlock.Stmts, a.analyzeLocal(stmt))
	}
	return hirBlock
}

func (a *Analyzer) analyzeLocal(local ast.Local) hir.Local {
	switch local := local.(type) {
	case *ast.Block:
		return a.analyzeBlock(local)
	case *ast.Return:
		return a.analyzeReturn(local)
	case ast.Expr:
		return a.analyzeExpr(local)
	case *ast.Let:
		return a.analyzeLet(local, false)
	default:
		panic("unreachable")
	}
}

func (a *Analyzer) analyzeReturn(ret *ast.Return) *hir.Return {
	ls := a.scope.(hir.LocalScope)

	var value optional.Optional[hir.Expr]
	if v, ok := ret.Value.Value(); ok {
		value = optional.Some(a.expectTypeExpr(v, ls.FuncType().Return))
	}
	return &hir.Return{
		Value: value,
	}
}

func (a *Analyzer) analyzeLet(l *ast.Let, isGlobal bool) *hir.Let {
	var value hir.Expr
	if isGlobal && l.Name.OriginText == "main" { // TODO: 同一个包下只允许存在一个main函数
		value = a.expectTypeExpr(l.Value, hir.NewFuncType(hir.Unit))
		if l.Mut {
			a.reporter.Fatalf(
				l.Name.Position,
				report.Errors.MustImmutable,
			)
		}
	} else {
		value = a.analyzeExpr(l.Value)
	}

	let := &hir.Let{
		Mut:   l.Mut,
		Name:  l.Name.OriginText,
		Value: value,
	}
	a.scope.AddValue(let)
	return let
}
