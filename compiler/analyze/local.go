package analyze

import (
	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/hir/scopes"
	"github.com/kkkunny/Sim/compiler/hir/stmts"
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/report"
)

func (a *Analyzer) analyzeBlock(block *ast.Block) *stmts.Block {
	hirBlock := stmts.NewBlock()
	for _, stmt := range block.Stmts {
		hirBlock.Stmts = append(hirBlock.Stmts, a.analyzeLocal(stmt))
	}
	return hirBlock
}

func (a *Analyzer) analyzeLocal(local ast.Local) stmts.Local {
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

func (a *Analyzer) analyzeReturn(ret *ast.Return) *stmts.Return {
	ls := a.scope.(scopes.LocalScope)
	if v, ok := ret.Value.Value(); ok {
		value := a.expectTypeExpr(v, ls.FuncType().Return)
		return stmts.NewReturn(value)
	} else {
		return stmts.NewReturn()
	}
}

func (a *Analyzer) analyzeLet(l *ast.Let, isGlobal bool) *stmts.Let {
	var t types.Type
	if tnode, ok := l.Type.Value(); ok {
		t = a.analyzeType(tnode)
	}

	var value stmts.Expr
	if isGlobal && l.Name.OriginText == "main" { // TODO: 同一个包下只允许存在一个main函数
		if l.Mut {
			a.reporter.Fatalf(
				l.Name.Position,
				report.Errors.MustImmutable,
			)
		}

		expectType := types.NewFuncType(types.Unit)
		v, ok := l.Value.Value()
		if !ok && t != nil {
			value = a.getZeroExpr(l.Name.Position, t)
		} else if t != nil {
			value = a.expectTypeExpr(v, expectType)
		} else {
			value = a.analyzeExpr(v, expectType)
		}
		if vt := value.GetType(); !vt.Equal(expectType) {
			a.reporter.Fatalf(
				l.Type.MustValue().Position(),
				report.Errors.UnexpectedExpression,
				expectType, vt,
			)
		}
	} else if v, ok := l.Value.Value(); t != nil && ok {
		value = a.expectTypeExpr(v, t)
	} else if t != nil {
		value = a.getZeroExpr(l.Name.Position, t)
	} else {
		value = a.analyzeExpr(v)
	}

	let := &stmts.Let{
		Mut:   l.Mut,
		Type:  value.GetType(),
		Name:  l.Name.OriginText,
		Value: value,
	}
	a.scope.AddValue(let)
	return let
}
