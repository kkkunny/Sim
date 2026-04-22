package analyze

import (
	"github.com/kkkunny/stl/container/either"

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
	case *ast.If:
		return a.analyzeIf(local)
	default:
		panic("unreachable")
	}
}

func (a *Analyzer) analyzeReturn(local *ast.Return) *stmts.Return {
	ls := a.scope.(scopes.LocalScope)
	if v, ok := local.Value.Value(); ok {
		value := a.expectTypeExpr(v, ls.FuncType().Return)
		return stmts.NewReturn(value)
	} else {
		return stmts.NewReturn()
	}
}

func (a *Analyzer) analyzeLet(local *ast.Let, isGlobal bool) *stmts.Let {
	var t types.Type
	if tnode, ok := local.Type.Value(); ok {
		t = a.analyzeType(tnode)
	}

	var value stmts.Expr
	if isGlobal && local.Name.OriginText == "main" { // TODO: 同一个包下只允许存在一个main函数
		if local.Mut {
			a.reporter.Fatalf(
				local.Name.Position,
				report.Errors.MustImmutable,
			)
		}

		expectType := types.NewFuncType(types.Unit)
		v, ok := local.Value.Value()
		if !ok && t != nil {
			value = a.getZeroExpr(local.Name.Position, t)
		} else if t != nil {
			value = a.expectTypeExpr(v, expectType)
		} else {
			value = a.analyzeExpr(v, expectType)
		}
		if vt := value.GetType(); !vt.Equal(expectType) {
			a.reporter.Fatalf(
				local.Type.MustValue().Position(),
				report.Errors.UnexpectedExpression,
				expectType, vt,
			)
		}
	} else if v, ok := local.Value.Value(); t != nil && ok {
		value = a.expectTypeExpr(v, t)
	} else if t != nil {
		value = a.getZeroExpr(local.Name.Position, t)
	} else {
		value = a.analyzeExpr(v)
	}

	let := &stmts.Let{
		Mut:   local.Mut,
		Type:  value.GetType(),
		Name:  local.Name.OriginText,
		Value: value,
	}
	a.scope.AddValue(let)
	return let
}

func (a *Analyzer) analyzeIf(local *ast.If) *stmts.If {
	cond := a.expectTypeExpr(local.Condition, types.Bool)

	a.scope = scopes.NewBlockScope(a.scope)
	body := a.analyzeBlock(local.Body)
	a.scope, _ = a.scope.Parent()

	var next []either.Either[*stmts.If, *stmts.Block]
	if nextLocal, ok := local.Else.Value(); ok {
		if elseif, ok := nextLocal.TryLeft(); ok {
			next = append(next, either.Left[*stmts.If, *stmts.Block](a.analyzeIf(elseif)))
		} else {
			a.scope = scopes.NewBlockScope(a.scope)
			next = append(next, either.Right[*stmts.If, *stmts.Block](a.analyzeBlock(nextLocal.Right())))
			a.scope, _ = a.scope.Parent()
		}
	}

	return stmts.NewIf(cond, body, next...)
}
