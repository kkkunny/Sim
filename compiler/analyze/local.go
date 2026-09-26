package analyze

import (
	"github.com/kkkunny/stl/container/either"
	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/locals"
	"github.com/kkkunny/Sim/compiler/hir/scopes"
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/report"
)

func (a *Analyzer) analyzeBlock(block *ast.Block) *locals.Block {
	hirBlock := locals.NewBlock()
	for _, stmt := range block.Stmts {
		func() {
			defer a.recoverFromAbort()
			hirBlock.Stmts = append(hirBlock.Stmts, a.analyzeLocal(stmt))
		}()
	}
	return hirBlock
}

func (a *Analyzer) analyzeLocal(local ast.Local) locals.Local {
	switch local := local.(type) {
	case *ast.Block:
		return a.analyzeBlock(local)
	case *ast.Return:
		return a.analyzeReturn(local)
	case ast.Expr:
		return a.analyzeExpr(local)
	case *ast.Let:
		return a.analyzeLocalLet(local)
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

func (a *Analyzer) analyzeReturn(local *ast.Return) *locals.Return {
	ls := a.scope.(*scopes.BlockScope)
	ret := ls.FuncType().GetReturn()
	if v, ok := local.Value.Value(); ok {
		value := a.expectTypeExpr(v, ret)
		return locals.NewReturn(value)
	}
	// 无值 return 仅允许出现在 unit 函数中；非 unit 函数在此拒绝（F7）
	if !isInvalidType(ret) && !isUnitType(ret) {
		a.errorf(
			local.BeginPosition,
			report.Errors.MissingReturnValue,
			ret,
		)
	}
	return locals.NewReturn()
}

func (a *Analyzer) analyzeLocalLet(local *ast.Let) *locals.Let {
	var t hir.Type
	if tAst, ok := local.Type.Value(); ok {
		t = a.analyzeType(tAst)
	}

	var value optional.Optional[locals.Expr]
	if v, ok := local.Value.Value(); t != nil && ok {
		value = optional.Some(a.expectTypeExpr(v, t))
	} else if t == nil {
		value = optional.Some(a.analyzeExpr(v))
		t = value.MustValue().GetType()
	} else if extern := stlslices.Any(local.Attributes, func(_ int, a ast.Attribute) bool { return stlval.Is[*ast.Extern](a) }); !extern {
		value = optional.Some(a.getZeroExpr(local.Name.Position, t))
	}

	// unit 类型不能作为变量存储（F8）
	if !isInvalidType(t) && isUnitType(t) {
		a.errorf(
			local.Name.Position,
			report.Errors.UnitTypedVariable,
			local.Name.OriginText,
		)
		t = types.Invalid
	}

	let := &locals.Let{
		Pub:   local.Public,
		Mut:   local.Mut,
		Name:  local.Name.OriginText,
		Type:  t,
		Value: value,
	}
	a.scope.AddValue(let)
	return let
}

func (a *Analyzer) analyzeIf(local *ast.If) *locals.If {
	cond := a.expectTypeExpr(local.Condition, types.Bool)

	prevScope := a.scope
	a.scope = scopes.NewBlockScope(prevScope)
	defer func() {
		a.scope = prevScope
	}()

	body := a.analyzeBlock(local.Body)
	a.scope = prevScope

	var next []either.Either[*locals.If, *locals.Block]
	if nextLocal, ok := local.Else.Value(); ok {
		if elseif, ok := nextLocal.TryLeft(); ok {
			next = append(next, either.Left[*locals.If, *locals.Block](a.analyzeIf(elseif)))
		} else {
			a.scope = scopes.NewBlockScope(prevScope)
			next = append(next, either.Right[*locals.If, *locals.Block](a.analyzeBlock(nextLocal.Right())))
			a.scope = prevScope
		}
	}

	return locals.NewIf(cond, body, next...)
}

func (a *Analyzer) analyzeWhile(local *ast.While) *locals.While {
	var cond locals.Expr
	if condAst, ok := local.Condition.Value(); ok {
		cond = a.expectTypeExpr(condAst, types.Bool)
	} else {
		cond = locals.NewBoolean(types.Bool, true)
	}

	prevScope := a.scope
	a.scope = scopes.NewBlockScope(prevScope)
	defer func() {
		a.scope = prevScope
	}()
	body := a.analyzeBlock(local.Body)
	a.scope = prevScope

	return locals.NewWhile(cond, body)
}

func (a *Analyzer) analyzeFor(local *ast.For) *locals.For {
	rangv := a.analyzeExpr(local.Range)
	et := hir.Type(types.Invalid)
	if at, ok := rangv.GetType().(types.ArrayType); ok {
		et = at.GetElem()
	} else if !isInvalidType(rangv.GetType()) {
		a.errorf(
			local.Range.Position(),
			report.Errors.UnexpectedExpressionCategory,
			"array", rangv.GetType(),
		)
	}
	param := hir.NewParam(local.Mut, et, local.Variable.OriginText)

	prevScope := a.scope
	a.scope = scopes.NewBlockScope(prevScope)
	defer func() {
		a.scope = prevScope
	}()
	a.scope.AddValue(param)
	body := a.analyzeBlock(local.Body)
	a.scope = prevScope

	return locals.NewFor(param, rangv, body)
}
