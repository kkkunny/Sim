package analyze

import (
	"github.com/kkkunny/stl/container/either"
	stlval "github.com/kkkunny/stl/value"

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
	case *ast.While:
		return a.analyzeWhile(local)
	case *ast.For:
		return a.analyzeFor(local)
	default:
		panic("unreachable")
	}
}

func (a *Analyzer) analyzeReturn(local *ast.Return) *stmts.Return {
	ls := a.scope.(scopes.LocalScope)
	if v, ok := local.Value.Value(); ok {
		value := a.expectTypeExpr(v, ls.FuncType().GetReturn())
		return stmts.NewReturn(value)
	} else {
		return stmts.NewReturn()
	}
}

func (a *Analyzer) analyzeLet(local *ast.Let, isGlobal bool) *stmts.Let {
	if isGlobal && local.Name.OriginText == "main" && local.Mut {
		a.reporter.Fatalf(
			local.Name.Position,
			report.Errors.MustImmutable,
		)
	}

	if isGlobal {
		_, ok := a.scope.Lookup(local.Name.OriginText)
		if ok {
			a.reporter.Fatalf(
				local.Name.Position,
				report.Errors.RepeatedIdentifier,
				local.Name.OriginText,
			)
		}
	}

	var t types.Type
	if tnode, ok := local.Type.Value(); ok {
		t = a.analyzeType(tnode)
	}

	let := stlval.IfLazy(
		isGlobal && local.Value.IsSome() && stlval.Is[*ast.Func](local.Value.MustValue()),
		func() *stmts.Let { // 允许自引用的let
			v := local.Value.MustValue().(*ast.Func)
			ft := a.analyzeFuncDecl(v)

			let := &stmts.Let{
				Mut:  local.Mut,
				Type: ft,
				Name: local.Name.OriginText,
			}
			a.scope.AddValue(let)

			let.Value = stlval.IfLazy(t == nil, func() stmts.Expr {
				return a.analyzeExpr(v)
			}, func() stmts.Expr {
				return a.expectTypeExpr(v, t)
			})
			return let
		},
		func() *stmts.Let {
			var value stmts.Expr
			if v, ok := local.Value.Value(); t != nil && ok {
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
		},
	)

	if isGlobal && local.Name.OriginText == "main" {
		expectType := types.NewFuncType(types.Unit)
		if vt := let.GetType(); !vt.Equal(expectType) {
			a.reporter.Fatalf(
				local.Type.MustValue().Position(),
				report.Errors.UnexpectedExpression,
				expectType, vt,
			)
		}
	}

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

func (a *Analyzer) analyzeWhile(local *ast.While) *stmts.While {
	var cond stmts.Expr
	if condAst, ok := local.Condition.Value(); ok {
		cond = a.expectTypeExpr(condAst, types.Bool)
	} else {
		cond = stmts.NewBoolean(true)
	}

	a.scope = scopes.NewBlockScope(a.scope)
	body := a.analyzeBlock(local.Body)
	a.scope, _ = a.scope.Parent()

	return stmts.NewWhile(cond, body)
}

func (a *Analyzer) analyzeFor(local *ast.For) *stmts.For {
	rangv := expectTypeExpr[types.ArrayType](a, local.Range)
	et := rangv.GetType().(types.ArrayType).GetElem()
	param := stmts.NewParam(local.Mut, et, local.Variable.OriginText)

	a.scope = scopes.NewBlockScope(a.scope)
	body := a.analyzeBlock(local.Body)
	a.scope, _ = a.scope.Parent()

	return stmts.NewFor(param, rangv, body)
}
