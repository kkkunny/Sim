package analyze

import (
	"github.com/kkkunny/stl/container/either"
	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/hir/scopes"
	"github.com/kkkunny/Sim/compiler/hir/stmts"
	"github.com/kkkunny/Sim/compiler/hir/types"
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
		return a.analyzeLetDef(local, false)
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

func (a *Analyzer) analyzeLetDef(local *ast.Let, isGlobal bool) *stmts.Let {
	var t types.Type
	if tAst, ok := local.Type.Value(); ok {
		t = a.analyzeType(tAst)
	}

	var value optional.Optional[stmts.Expr]
	if v, ok := local.Value.Value(); t != nil && ok {
		value = optional.Some(a.expectTypeExpr(v, t))
	} else if t == nil {
		value = optional.Some(a.analyzeExpr(v))
		t = value.MustValue().GetType()
	} else if extern := stlslices.Any(local.Attributes, func(_ int, a ast.Attribute) bool { return stlval.Is[*ast.Extern](a) }); !extern {
		value = optional.Some(a.getZeroExpr(local.Name.Position, t))
	}

	var let *stmts.Let
	if decl, ok := a.scope.LookupValue(local.Name.OriginText); ok && isGlobal && stlval.Is[*stmts.Let](decl) && decl.(*stmts.Let).Global {
		let = decl.(*stmts.Let)
	} else {
		let = &stmts.Let{
			Pub:    local.Public,
			Global: isGlobal,
			Mut:    local.Mut,
			Type:   t,
			Name:   local.Name.OriginText,
		}
		a.scope.AddValue(let)
	}
	let.Value = value
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
		cond = stmts.NewBoolean(types.Bool, true)
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
